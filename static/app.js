const chatBox = document.getElementById('chat-box');
const inputZone = document.getElementById('input-zone');
const msgInput = document.getElementById('msg-input');

// WebSocket connection
// Use current host/port, but switch protocol to ws/wss
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = `${protocol}//${window.location.host}/ws`;

console.log(`Connecting to WebSocket at ${wsUrl}...`);
const socket = new WebSocket(wsUrl);

socket.onopen = () => {
    console.log('WebSocket Connection Open');
    addSystemMessage('Connected to chat server!');
};

socket.onmessage = (event) => {
    try {
        const msg = JSON.parse(event.data);
        addMessage(msg.sender_id, msg.content);
    } catch (e) {
        console.error('Error parsing message:', e);
    }
};

socket.onclose = (event) => {
    console.log('WebSocket Connection Closed', event);
    addSystemMessage('Disconnected from server.');
};

socket.onerror = (error) => {
    console.error('WebSocket Error:', error);
    addSystemMessage('Connection error.');
};

// Handle form submission
inputZone.addEventListener('submit', (e) => {
    e.preventDefault();
    const content = msgInput.value.trim();

    if (content && socket.readyState === WebSocket.OPEN) {
        const payload = { content: content };
        socket.send(JSON.stringify(payload));
        msgInput.value = '';
    } else if (socket.readyState !== WebSocket.OPEN) {
        addSystemMessage('Cannot send message: no connection.');
    }
});

// UI Helper functions
function addMessage(sender, content) {
    const msgDiv = document.createElement('div');
    msgDiv.classList.add('message');

    // Check if it's "me" or someone else (simple heuristic for now)
    // The server currently sends back UUIDs, we don't know our own UUID from the client side easily in this setup 
    // unless we parse the initial message or handshake. For now, just display Sender ID.

    const senderDiv = document.createElement('div');
    senderDiv.classList.add('sender');
    senderDiv.textContent = sender; // Currently UUID

    const contentDiv = document.createElement('div');
    contentDiv.classList.add('content');
    contentDiv.textContent = content;

    msgDiv.appendChild(senderDiv);
    msgDiv.appendChild(contentDiv);

    chatBox.appendChild(msgDiv);
    chatBox.scrollTop = chatBox.scrollHeight;
}

function addSystemMessage(text) {
    const msgDiv = document.createElement('div');
    msgDiv.classList.add('system-msg');
    msgDiv.textContent = text;
    chatBox.appendChild(msgDiv);
    chatBox.scrollTop = chatBox.scrollHeight;
}
