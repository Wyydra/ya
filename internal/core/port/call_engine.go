package port

import "github.com/Wyydra/ya/internal/core/domain"

type CallEngine interface {
	ProcessJoin(neg domain.CallNegotiation) (domain.CallNegotiation, error)
	AddNetworkRoute(neg domain.CallNegotiation) error
	TerminateCall(userID domain.UserID) error
}
