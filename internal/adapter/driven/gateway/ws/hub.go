package ws

import (
	"context"
	"errors"
	"sync" // Added sync

	"github.com/Wyydra/ya/internal/core/domain"
	"github.com/rs/zerolog/log"
)

// implements port.RealTimeGateway
type Hub struct {
	mu         sync.Mutex
	clients    map[Client]bool
	broadcast  chan domain.Message
	register   chan Client
	unregister chan Client
	quit       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[Client]bool),
		broadcast:  make(chan domain.Message),
		register:   make(chan Client),
		unregister: make(chan Client),
		quit:       make(chan struct{}),
	}
}

func (h *Hub) BroadcastMessage(ctx context.Context, msg domain.Message) error {
	select {
	case h.broadcast <- msg:
	default:
		log.Warn().Msg("Broadcast channel full, dropping message")
	}
	return nil
}

func (h *Hub) SendCallSignal(ctx context.Context, userID domain.UserID, negotiation domain.CallNegotiation) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if client.ID() == userID.String() {
			return client.SendCall(negotiation)
		}
	}
	return nil // Client not found, maybe offline, ignore or error
}

func (h *Hub) NotifyUserJoined(ctx context.Context, roomID domain.RoomID, userID domain.UserID) error {
	return errors.New("not implemented")
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.quit:
			for client := range h.clients {
				client.Close()
				delete(h.clients, client)
			}
			return

		case client := <-h.register:
			h.clients[client] = true
			log.Info().Str("client_id", client.ID()).Msg("Client registered")

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Info().Str("client_id", client.ID()).Msg("Client unregistered")
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				// TODO: check if client is in the same room as message.RoomID
				if err := client.SendText(message); err != nil {
					log.Error().Err(err).Str("client_id", client.ID()).Msg("Error sending message")
					client.Close()
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) Register(c Client) {
	h.register <- c
}

func (h *Hub) Unregister(c Client) {
	h.unregister <- c
}

func (h *Hub) Stop() {
	close(h.quit)
}
