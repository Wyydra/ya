package service

import (
	"context"

	"github.com/Wyydra/ya/backend/internal/core/domain"
	"github.com/Wyydra/ya/backend/internal/core/port"
	"github.com/rs/zerolog/log"
)

type CallService struct {
	media  port.MediaEngine
	gateway port.RealTimeGateway
}

func NewCallService(media port.MediaEngine, gateway port.RealTimeGateway) *CallService {
	s := &CallService{
		media:   media,
		gateway: gateway,
	}
	
	media.SetSignalCallback(func(sessionID domain.SessionID, userID domain.UserID, signal domain.Signal) {
		// TODO: context ?
		if err := gateway.SendSignal(context.Background(), userID, signal); err != nil {
			log.Error().Err(err).
				Str("sessionID", sessionID.String()).
				Str("userID", userID.String()).
				Msg("failed to send signal to gateway")
		}
	})
	
	return s
}

func (s *CallService) JoinCall(ctx context.Context, roomID domain.RoomID, userID domain.UserID) error {
	// map RoomID -> SesssionID //TODO: is it good?
	sessionID := domain.SessionID(roomID.String())
	
	offer, err := s.media.AddPeer(sessionID, userID)
	if err != nil {
		return err
	}

	return s.gateway.SendSignal(ctx, userID, offer)
}

func (s *CallService) HandleSignal(ctx context.Context, userID domain.UserID, roomID domain.RoomID, signal domain.Signal) error {
	sessionID := domain.SessionID(roomID.String()) //TODO is it good ?
	return s.media.HandleSignal(sessionID, userID, signal)
}

func (s *CallService) LeaveCall(ctx context.Context, roomID domain.RoomID, userID domain.UserID) error {
	sessionID := domain.SessionID(roomID.String())
	s.media.RemovePeer(sessionID, userID)
	return nil
}
