package ws

import "github.com/Wyydra/ya/backend/internal/core/domain"

type Client interface {
	ID() string
	SendText(msg domain.Message) error
	SendSignal(signal domain.Signal) error
	Close() error
}
