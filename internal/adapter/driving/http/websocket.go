package http

import (
	"net/http"

	"github.com/Wyydra/ya/internal/core/domain"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
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

func (c* WSClient) SendCall(neg domain.CallNegotiation) error {
	return c.conn.WriteJSON(map[string]interface{}{
		"event": "call_signal",
		"intent": neg.Intent,
		"paylaod": string(neg.Payload),
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
		conn.Close()
	}()

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

		// TODO: Parse RoomID from request or context
		roomID := domain.NewRoomID() 

		if req.Type == "call_signal" {
			neg := domain.CallNegotiation{
				UserID:  client.id,
				RoomID:  roomID,
				Intent:  domain.CallIntent(req.Intent), // Cast string to CallIntent
				Payload: []byte(req.Payload),
			}
			if err := h.CallService.HandleSignal(r.Context(), neg); err != nil {
				l.Error().Err(err).Msg("Failed to handle call signal")
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
