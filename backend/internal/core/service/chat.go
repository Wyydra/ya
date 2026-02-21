package service

import (
	"context"

	"github.com/Wyydra/ya/backend/internal/core/domain"
	"github.com/Wyydra/ya/backend/internal/core/port"
)

type ChatService struct {
	repo    port.MessageRepository
	gateway port.RealTimeGateway
}

func NewChatService(repo port.MessageRepository, gateway port.RealTimeGateway) *ChatService {
	return &ChatService{
		repo:    repo,
		gateway: gateway,
	}
}

func (s *ChatService) SendMessage(ctx context.Context, senderID domain.UserID, roomID domain.RoomID, content string) error {
	msg, err := domain.NewMessage(senderID, roomID, content)
	if err != nil {
		return err
	}

	if err := s.repo.Save(ctx, *msg); err != nil {
		return err
	}
	return s.gateway.BroadcastMessage(ctx, *msg)
}
