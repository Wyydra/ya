package service

import (
	"context"

	"github.com/Wyydra/ya/internal/core/domain"
	"github.com/Wyydra/ya/internal/core/port"
	"github.com/rs/zerolog/log"
)

type CallService struct {
	engine  port.CallEngine
	gateway port.RealTimeGateway
}

func NewCallService(engine port.CallEngine, gateway port.RealTimeGateway) *CallService {
	return &CallService{
		engine:  engine,
		gateway: gateway,
	}
}

func (s *CallService) HandleSignal(ctx context.Context, neg domain.CallNegotiation) error {
	switch neg.Intent {
	case domain.IntentJoin:
		response, err := s.engine.ProcessJoin(neg)
		if err != nil {
			log.Err(err).Msg("CallEngine error")
			return err
		}
		
		return s.gateway.SendCallSignal(ctx, neg.UserID, response) // Send back to caller
	case domain.IntentNetwork:
		return s.engine.AddNetworkRoute(neg)
	default:
		return nil
	}
}

func (s *CallService) LeaveCall(ctx context.Context, userID domain.UserID) error {
	return s.engine.TerminateCall(userID)
}
