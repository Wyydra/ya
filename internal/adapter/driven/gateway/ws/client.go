package ws

import "github.com/Wyydra/ya/internal/core/domain"

type Client interface {
	ID() string
	SendText(msg domain.Message) error
	SendCall(neg domain.CallNegotiation) error
	Close() error
}
