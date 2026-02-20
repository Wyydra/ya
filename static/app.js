// app.js

class YaClient {
    constructor() {
        this.socket = null;
        this.pc = null;
        this.localStream = null;
        this.isVoiceConnected = false;

        // UI References
        this.ui = {
            messages: document.getElementById('messages'),
            input: document.getElementById('chat-input'),
            form: document.getElementById('chat-form'),
            joinBtn: document.getElementById('join-btn'),
            videoGrid: document.getElementById('video-section'),
            localVideo: document.getElementById('local-video'),
        };

        this.bindEvents();
        this.connectWS();
    }

    bindEvents() {
        this.ui.form.addEventListener('submit', (e) => {
            e.preventDefault();
            const text = this.ui.input.value.trim();
            if (text) {
                this.sendJSON({ content: text }); // Chat message
                this.ui.input.value = '';
            }
        });

        this.ui.joinBtn.addEventListener('click', () => {
            if (!this.isVoiceConnected) {
                this.joinVoice();
            } else {
                console.log("Already connected (Leave not implemented yet)");
            }
        });
    }

    // --- WebSocket ---

    connectWS() {
        const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const url = `${proto}//${window.location.host}/ws`;

        console.log(`Connecting to ${url}`);
        this.socket = new WebSocket(url);

        this.socket.onopen = () => {
            this.logSystem('Connected to server via WebSocket.');
        };

        this.socket.onclose = () => {
            this.logSystem('Disconnected from server.');
        };

        this.socket.onerror = (err) => {
            console.error("WS Error:", err);
            this.logSystem('WebSocket connection error.');
        };

        this.socket.onmessage = (e) => this.handleMessage(e);
    }

    handleMessage(event) {
        try {
            const msg = JSON.parse(event.data);

            if (msg.type === 'signal') {
                if (!msg.payload) {
                    console.error("Signal payload is missing! Server might be sending wrong key.");
                    return;
                }
                this.handleSignal(msg.payload);
            } else if (msg.sender_id) {
                this.addChatMessage(msg.sender_id, msg.content);
            }
        } catch (err) {
            console.error("Failed to parse message:", event.data, err);
        }
    }

    sendJSON(obj) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(obj));
        } else {
            this.logSystem("Cannot send: disconnected.");
        }
    }

    // --- WebRTC Core ---

    async joinVoice() {
        this.logSystem("Requesting microphone/camera access...");
        try {
            // 1. Get User Media
            try {
                this.localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
            } catch (e) {
                console.warn("Video failed, trying audio only", e);
                this.localStream = await navigator.mediaDevices.getUserMedia({ video: false, audio: true });
            }

            // Display Local Video
            this.ui.localVideo.srcObject = this.localStream;
            this.isVoiceConnected = true;
            this.ui.joinBtn.textContent = "Voice Active";
            this.ui.joinBtn.classList.replace('btn-green', 'btn-red');

            // 2. Send 'join_call' to server => Server sends Offer
            this.sendJSON({ type: "join_call" });
            this.logSystem("Joining call...");

        } catch (err) {
            console.error("Media Error:", err);
            this.logSystem(`Could not access media devices: ${err.name}`);
        }
    }

    async handleSignal(signal) {
        console.log("Received Signal:", signal.Type);

        if (!this.pc) this.createPeerConnection();

        try {
            switch (signal.Type) {
                case 'offer':
                    await this.handleOffer(signal.Payload);
                    break;
                case 'candidate':
                    await this.handleCandidate(signal.Payload);
                    break;
                case 'answer':
                    // We are usually the answerer in this flow (Server offers), 
                    // but if we were offering, we'd handle answer here.
                    await this.pc.setRemoteDescription({ type: 'answer', sdp: signal.Payload });
                    break;
            }
        } catch (err) {
            console.error("Signal Handling Error:", err);
        }
    }

    createPeerConnection() {
        this.pc = new RTCPeerConnection({
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' }
            ]
        });

        // 1. Add Local Tracks
        this.localStream.getTracks().forEach(track => {
            this.pc.addTrack(track, this.localStream);
        });

        // 2. Handle Remote Tracks (Create video elements)
        this.pc.ontrack = (event) => {
            const stream = event.streams[0];
            console.log("OnTrack:", event.track.kind, stream.id);

            // Avoid duplicates
            let vid = document.getElementById(`vid-${stream.id}`);
            if (!vid) {
                vid = document.createElement('video');
                vid.id = `vid-${stream.id}`;
                vid.autoplay = true;
                vid.playsInline = true;

                // Hack: Sometimes webkit needs controls or interaction
                // vid.controls = true; 

                this.ui.videoGrid.appendChild(vid);
            }
            vid.srcObject = stream;

            // Helper to update audio-only styling
            const updateAudioOnlyState = () => {
                const videoTracks = stream.getVideoTracks();
                if (videoTracks.length === 0) {
                    vid.classList.add('audio-only');
                } else {
                    vid.classList.remove('audio-only');
                }

                // If no tracks left at all (audio or video), remove the element
                if (stream.getTracks().length === 0) {
                    console.log("Stream has no tracks left, removing element:", stream.id);
                    const el = document.getElementById(`vid-${stream.id}`);
                    if (el) el.remove();
                }
            };

            // Initial check
            updateAudioOnlyState();

            // Listen for track changes
            stream.onaddtrack = (e) => {
                console.log("Stream addtrack:", e.track.kind);
                updateAudioOnlyState();
            };

            stream.onremovetrack = (e) => {
                console.log("Stream removetrack:", e.track.kind);
                updateAudioOnlyState();
            };

            // Fallback: Clean up when track ends explicitly
            event.track.onended = () => {
                console.log("Track ended:", event.track.kind);
                updateAudioOnlyState();
            };
        };

        // 3. Handle ICE Candidates
        this.pc.onicecandidate = (event) => {
            if (event.candidate) {
                this.sendSignal('candidate', JSON.stringify(event.candidate));
            }
        };

        // 4. Connection State Monitoring
        this.pc.onconnectionstatechange = () => {
            console.log("PC State:", this.pc.connectionState);
        };
    }

    async handleOffer(sdp) {
        await this.pc.setRemoteDescription({ type: 'offer', sdp: sdp });
        const answer = await this.pc.createAnswer();
        await this.pc.setLocalDescription(answer);
        this.sendSignal('answer', answer.sdp);
    }

    async handleCandidate(candidateJSON) {
        if (!candidateJSON) return;
        const candidate = JSON.parse(candidateJSON);
        await this.pc.addIceCandidate(candidate);
    }

    sendSignal(type, payload) {
        // Wrap in the format expected by the backend
        // Backend expects: { type: "signal", payload: "{ type: 'answer', payload: '...' }" }
        const innerPayload = JSON.stringify({
            type: type,
            payload: payload
        });

        this.sendJSON({
            type: "signal",
            payload: innerPayload
        });
    }

    // --- UI Helpers ---

    addChatMessage(sender, text) {
        const div = document.createElement('div');
        div.className = 'message';
        div.innerHTML = `<div class="author">${sender}</div><div class="content">${this.escapeHtml(text)}</div>`;
        this.ui.messages.appendChild(div);
        this.ui.messages.scrollTop = this.ui.messages.scrollHeight;
    }

    logSystem(text) {
        const div = document.createElement('div');
        div.className = 'system-msg';
        div.textContent = text;
        this.ui.messages.appendChild(div);
        this.ui.messages.scrollTop = this.ui.messages.scrollHeight;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Start
window.addEventListener('DOMContentLoaded', () => {
    window.app = new YaClient();
});
