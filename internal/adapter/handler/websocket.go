package handler

import (
	"net/http"

	"github.com/Wyydra/ya/internal/core/domain"
	"github.com/google/uuid"
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
	id string
	conn *websocket.Conn
}

func (c *WSClient) ID() string {
	return c.id
}

func (c *WSClient) Send(msg domain.Message) error {
	type messageDTO struct {
		SenderID string `json:"sender_id"`
		Content string `json:"content"`
	}

	dto := messageDTO {
		SenderID: msg.SenderID,
		Content: msg.Content,
	}

	return c.conn.WriteJSON(dto)
}

func (c *WSClient) Close() error {
	return c.conn.Close()
}

// HTTP handler
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error while upgrading ws")
		return
	}

	clientID := uuid.New().String()
	
	client := &WSClient{
		id:   clientID,
		conn: conn,
	}

	// Contextual logging for this connection
	l := log.With().Str("client_id", clientID).Logger()
	l.Info().Msg("New client connected")

	h.RoomService.Join(client)

	defer func() {
		l.Info().Msg("Client disconnected")
		h.RoomService.Leave(client)
		conn.Close()
	}()

	// listening for browser
	for {
		type incomingDTO struct {
			Content string `json:"content"`
		}

		var req incomingDTO
		err := conn.ReadJSON(&req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				l.Error().Err(err).Msg("Unexpected close error")
			}
			break 
		}

		domainMsg := domain.Message{
			SenderID: client.ID(),
			Content:  req.Content,
		}
		
		h.RoomService.BroadcastMessage(domainMsg)
	}
}
