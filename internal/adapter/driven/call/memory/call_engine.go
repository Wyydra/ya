package memory

import (
	"github.com/Wyydra/ya/internal/core/domain"
)

type CallEngine struct{}

func NewCallEngine() *CallEngine {
	return &CallEngine{}
}

func (e *CallEngine) ProcessJoin(neg domain.CallNegotiation) (domain.CallNegotiation, error) {
	// Stub: just echo back successfully
	return neg, nil
}

func (e *CallEngine) AddNetworkRoute(neg domain.CallNegotiation) error {
	return nil
}

func (e *CallEngine) TerminateCall(userID domain.UserID) error {
	return nil
}
