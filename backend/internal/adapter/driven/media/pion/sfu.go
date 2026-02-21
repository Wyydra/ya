package pion

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/Wyydra/ya/backend/internal/core/domain"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
)

type Peer struct {
	ID domain.UserID
	PC *webrtc.PeerConnection
	
	mu sync.Mutex
	negotiationPending bool // True if we need to renegotiate but were in unstable state
}

type trackInfo struct {
	Track *webrtc.TrackLocalStaticRTP
	Owner domain.UserID
}

type PionAdapter struct {
	api *webrtc.API
	// SessionID -> UserID -> Peer
	sessions map[domain.SessionID]map[domain.UserID]*Peer
	// SessionID -> List of Tracks in that session
	tracks map[domain.SessionID][]trackInfo
	mu     sync.RWMutex
	
	onSignal func(sessionID domain.SessionID, userID domain.UserID, signal domain.Signal)
}

func NewPionAdapter() *PionAdapter {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	return &PionAdapter{
		api:      api,
		sessions: make(map[domain.SessionID]map[domain.UserID]*Peer),
		tracks:   make(map[domain.SessionID][]trackInfo),
	}
}

func (a *PionAdapter) SetSignalCallback(cb func(sessionID domain.SessionID, userID domain.UserID, signal domain.Signal)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onSignal = cb
}

func (a *PionAdapter) PeerID(p *Peer) domain.UserID { return p.ID }

func (a *PionAdapter) AddPeer(sessionID domain.SessionID, userID domain.UserID) (domain.Signal, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Initialize session if needed
	if _, ok := a.sessions[sessionID]; !ok {
		a.sessions[sessionID] = make(map[domain.UserID]*Peer)
		a.tracks[sessionID] = []trackInfo{}
	}

	// Create Peer Connection
	pc, err := a.api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return domain.Signal{}, err
	}

	// Important: Add Transceivers acting as "RecvOnly" 
	// to tell the client "I am ready to receive Audio and Video".
	// This ensures the Offer has m=audio and m=video sections.
	// We use RecvOnly because we don't have any specific track to send on *these* transceivers yet.
	// Future tracks added via AddTrack will create their own transceivers (SendOnly).
	if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		return domain.Signal{}, err
	}
	if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		return domain.Signal{}, err
	}

	peer := &Peer{ID: userID, PC: pc}
	a.sessions[sessionID][userID] = peer

	// 1. EVENT: Allow Trickle ICE
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		
		candidateJSON, err := json.Marshal(c.ToJSON())
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal candidate")
			return
		}
		
		signal := domain.NewSignal(domain.SignalCandidate, string(candidateJSON))
		
		a.mu.RLock()
		cb := a.onSignal
		a.mu.RUnlock()
		
		if cb != nil {
			cb(sessionID, userID, signal)
		}
	})

	// 2. EVENT: When this peer sends a track (Forward it to others)
	pc.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Debug().Str("kind", remoteTrack.Kind().String()).Str("user_id", userID.String()).Msg("Received remote track")
		
		// Create a local track to relay media
		localTrack, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, remoteTrack.ID(), remoteTrack.StreamID())
		if err != nil {
			log.Error().Err(err).Msg("Failed to create local track")
			return
		}

		// Save track definition
		a.mu.Lock()
		a.tracks[sessionID] = append(a.tracks[sessionID], trackInfo{Track: localTrack, Owner: userID})
		
		// Add this new track to ALL OTHER existing peers
		for otherID, otherPeer := range a.sessions[sessionID] {
			if otherID != userID {
				// Safety check
				if otherPeer.PC.ConnectionState() == webrtc.PeerConnectionStateClosed {
					continue
				}
				
				if _, err := otherPeer.PC.AddTrack(localTrack); err != nil {
					log.Error().Err(err).Msg("Failed to add track to other peer")
				} else {
					// Renegotiate!
					go a.renegotiate(sessionID, otherID, otherPeer)
				}
			}
		}
		a.mu.Unlock()

		// Start relay loop
		go func() {
			rtpBuf := make([]byte, 1400)
			for {
				i, _, err := remoteTrack.Read(rtpBuf)
				if err != nil {
					if err == io.EOF { return }
					return
				}
				if _, err := localTrack.Write(rtpBuf[:i]); err != nil {
					if err == io.EOF { return }
					return
				}
			}
		}()

		// Send PLI (Picture Loss Indication) every 3 seconds AND immediately
		go func() {
			sendPLI := func() {
				if rtcpErr := pc.WriteRTCP([]rtcp.Packet{
					&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
				}); rtcpErr != nil {
					// Benign error on closed connection
				}
			}
			
			// Send immediate PLI to request keyframe ASAP
			sendPLI()

			ticker := time.NewTicker(time.Second * 3)
			defer ticker.Stop()
			for range ticker.C {
				sendPLI()
			}
		}()
	})

	// 3. Add EXISTING tracks to this new peer
	for _, t := range a.tracks[sessionID] {
		if t.Owner != userID { // Don't send back own video
			if _, err := pc.AddTrack(t.Track); err != nil {
				log.Error().Err(err).Msg("Failed to add existing track to new peer")
			}
		}
	}

	// 4. Create Offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return domain.Signal{}, err
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		return domain.Signal{}, err
	}

	// Wait briefly for gathering to start
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	
	done := webrtc.GatheringCompletePromise(pc)
	select {
	case <-done:
	case <-ctx.Done():
	}

	return domain.NewSignal(domain.SignalOffer, pc.LocalDescription().SDP), nil
}

func (a *PionAdapter) renegotiate(sessionID domain.SessionID, userID domain.UserID, peer *Peer) {
	peer.mu.Lock()
	defer peer.mu.Unlock()
	
	if peer.PC.ConnectionState() == webrtc.PeerConnectionStateClosed {
		return
	}

	// Check signaling state
	if peer.PC.SignalingState() != webrtc.SignalingStateStable {
		log.Debug().Str("user_id", userID.String()).Msg("Renegotiation: Signaling state not stable, queuing")
		peer.negotiationPending = true
		return
	}

	offer, err := peer.PC.CreateOffer(nil)
	if err != nil {
		log.Error().Err(err).Msg("Renegotiation: Failed to create offer")
		return
	}
	
	if err := peer.PC.SetLocalDescription(offer); err != nil {
		log.Error().Err(err).Msg("Renegotiation: Failed to set local description")
		return
	}
	
	signal := domain.NewSignal(domain.SignalOffer, peer.PC.LocalDescription().SDP)
	
	a.mu.RLock()
	cb := a.onSignal
	a.mu.RUnlock()
	
	if cb != nil {
		cb(sessionID, userID, signal)
	}
}

func (a *PionAdapter) HandleSignal(sessionID domain.SessionID, userID domain.UserID, signal domain.Signal) error {
	a.mu.RLock()
	session, ok := a.sessions[sessionID]
	if !ok {
		a.mu.RUnlock()
		return errors.New("session not found")
	}
	peer, ok := session[userID]
	a.mu.RUnlock()
	
	if !ok {
		return errors.New("peer not found")
	}

	switch signal.Type {
	case domain.SignalAnswer:
		log.Debug().Int("sdp_len", len(signal.Payload)).Msg("Setting Remote Description (Answer)")
		
		sdp := webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: signal.Payload}
		if err := peer.PC.SetRemoteDescription(sdp); err != nil {
			return err
		}
		
		// If we had a pending negotiation, trigger it now (asynchronously to release lock)
		peer.mu.Lock()
		pending := peer.negotiationPending
		if pending {
			peer.negotiationPending = false
		}
		peer.mu.Unlock()
		
		if pending {
			log.Debug().Str("user_id", userID.String()).Msg("Triggering queued renegotiation")
			go a.renegotiate(sessionID, userID, peer)
		}
		return nil
		
	case domain.SignalCandidate:
		var candidate webrtc.ICECandidateInit
		if err := json.Unmarshal([]byte(signal.Payload), &candidate); err != nil {
			return err
		}
		return peer.PC.AddICECandidate(candidate)
	}
	return nil
}

func (a *PionAdapter) RemovePeer(sessionID domain.SessionID, userID domain.UserID) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	session, ok := a.sessions[sessionID]
	if !ok {
		return
	}
	
	// 1. Close the leaving peer
	if peer, ok := session[userID]; ok {
		peer.PC.Close()
		delete(session, userID)
	}

	// 2. Identify tracks to remove
	var remainingTracks []trackInfo
	var tracksToRemove []*webrtc.TrackLocalStaticRTP
	
	for _, t := range a.tracks[sessionID] {
		if t.Owner == userID {
			tracksToRemove = append(tracksToRemove, t.Track)
		} else {
			remainingTracks = append(remainingTracks, t)
		}
	}
	a.tracks[sessionID] = remainingTracks

	// 3. Remove these tracks from all other peers
	if len(tracksToRemove) > 0 {
		for otherID, otherPeer := range session {
			if otherPeer.PC.ConnectionState() == webrtc.PeerConnectionStateClosed {
				continue
			}

			needsRenegotiation := false
			
			senders := otherPeer.PC.GetSenders()
			for _, sender := range senders {
				track := sender.Track()
				if track == nil {
					continue
				}
				
				// Check if this track is one of the removed ones
				for _, removedTrack := range tracksToRemove {
					if track == removedTrack {
						if err := otherPeer.PC.RemoveTrack(sender); err != nil {
							log.Error().Err(err).Str("user_id", otherID.String()).Msg("Failed to remove track")
						} else {
							needsRenegotiation = true
						}
					}
				}
			}
			
			if needsRenegotiation {
				go a.renegotiate(sessionID, otherID, otherPeer)
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
