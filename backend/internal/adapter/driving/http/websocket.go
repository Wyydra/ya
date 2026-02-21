package http

import (
	"encoding/json"
	"net/http"

	"github.com/Wyydra/ya/backend/internal/core/domain"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const DemoRoomID = "db31e952-84dd-40c4-9bed-b7ddd35ba5b8"

var upgrader = websocket.Upgrader {
	ReadBufferSize: 4096,
	WriteBufferSize: 4096,
	// TODO: only for dev
	CheckOrigin: func(r *http.Request) bool {return true},
}

type WSClient struct {
	id domain.UserID
	conn *websocket.Conn
}

func (c *WSClient) ID() string {
	return c.id.String()
}

func (c *WSClient) SendText(msg domain.Message) error {
	type messageDTO struct {
		SenderID string `json:"sender_id"`
		Content string `json:"content"`
	}

	dto := messageDTO {
		SenderID: msg.SenderID.String(),
		Content: msg.Content,
	}

	return c.conn.WriteJSON(dto)
}

func (c *WSClient) Close() error {
	return c.conn.Close()
}

func (c* WSClient) SendSignal(signal domain.Signal) error {
	return c.conn.WriteJSON(map[string]interface{}{
		"type":    "signal",
		"payload": signal,
	})
}

// HTTP handler
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error while upgrading ws")
		return
	}

	clientID := domain.NewUserID()
	
	client := &WSClient{
		id:   clientID,
		conn: conn,
	}

	l := log.With().Str("client_id", clientID.String()).Logger()
	l.Info().Msg("New client connected")

	h.Hub.Register(client)

	defer func() {
		l.Info().Msg("Client disconnected")
		h.Hub.Unregister(client)
		
		// Cleanup SFU peer
		// Using the same hardcoded RoomID as below
		roomID, _ := domain.NewRoomIDFromString(DemoRoomID)
		if err := h.CallService.LeaveCall(r.Context(), roomID, client.id); err != nil {
             // benign error
        }
		
		conn.Close()
	}()

	// TODO: Parse RoomID from request or context
	// For now, use a constant RoomID so everyone joins the same room
	roomID, _ := domain.NewRoomIDFromString(DemoRoomID) // Use a fixed UUID for testing
	
	// listening for browser
	for {
		type incomingDTO struct {
			Type    string `json:"type"`
			Content string `json:"content"`
			Intent  string `json:"intent"`
			Payload string `json:"payload"`
		}

		var req incomingDTO
		err := conn.ReadJSON(&req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				l.Error().Err(err).Msg("Unexpected close error")
			}
			break 
		}

		type incomingSignalDTO struct {
			Type    string `json:"type"`    // "offer", "answer", "candidate"
			Payload string `json:"payload"` // The opaque SDP/ICE string
		}

		if req.Type == "signal" {
			var sigDTO incomingSignalDTO
			if err := json.Unmarshal([]byte(req.Payload), &sigDTO); err != nil {
				l.Error().Err(err).Msg("Invalid signal payload")
				continue
			}

			sig := domain.NewSignal(domain.SignalType(sigDTO.Type), sigDTO.Payload)

			if err := h.CallService.HandleSignal(r.Context(), client.id, roomID, sig); err != nil {
				l.Error().Err(err).Msg("Failed to handle signal")
			}

		} else if req.Type == "join_call" {
			// Trigger the JoinCall flow (Server will create Offer)
			if err := h.CallService.JoinCall(r.Context(), roomID, client.id); err != nil {
				l.Error().Err(err).Msg("Failed to join call")
			}
		} else {
			// Default to chat
			err = h.ChatService.SendMessage(r.Context(), client.id, roomID, req.Content)
			if err != nil {
				l.Error().Err(err).Msg("Failed to process message")
				continue
			}
		}
	}
}
