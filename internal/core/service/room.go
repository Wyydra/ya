package service

import (
	"github.com/Wyydra/ya/internal/core/domain"
	"github.com/Wyydra/ya/internal/core/port"
	"github.com/rs/zerolog/log"
)

type RoomService struct {
	clients map[port.Client]bool
	broadcast chan domain.Message
	register chan port.Client
	unregister chan port.Client
	quit chan struct{}
}

func NewRoomService() *RoomService {
	return &RoomService{
		clients: make(map[port.Client]bool),
		broadcast: make(chan domain.Message),
		register: make(chan port.Client),
		unregister: make(chan port.Client),
		quit: make(chan struct{}),
	}
}

func (s *RoomService) Join(c port.Client) {
	s.register <- c
}

func (s *RoomService) Leave(c port.Client) {
	s.unregister <- c
}

func (s *RoomService) BroadcastMessage(msg domain.Message) {
	s.broadcast <- msg
}

func (s *RoomService) Stop() {
	close(s.quit)
}

func (s *RoomService) Run() {
	for {
		select {
		case <- s.quit:
			log.Info().Msg("Stopping RoomService. Disconnecting all clients.")
			for client := range s.clients {
				if err := client.Close(); err != nil {
					log.Error().Err(err).Str("client_id", client.ID()).Msg("Error closing client connection")
				}
				delete(s.clients, client)
			}
			return

		case client := <- s.register:
			s.clients[client] = true
			log.Info().Int("count", len(s.clients)).Str("client_id", client.ID()).Msg("Client joined room")

		case client := <- s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				log.Info().Int("count", len(s.clients)).Str("client_id", client.ID()).Msg("Client left room")
			}

		case message := <- s.broadcast:
			log.Debug().Str("client_id", message.SenderID).Str("content", message.Content).Msg("New messsage")
			for client := range s.clients {
				err := client.Send(message)
				if err != nil {
					log.Error().Err(err).Str("client_id", client.ID()).Msg("Error broadcasting message")
					delete(s.clients, client)
				}
			}
		}
	}
}

