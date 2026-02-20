package ws

import "github.com/Wyydra/ya/internal/core/domain"

type Client interface {
	ID() string
	SendText(msg domain.Message) error
	SendSignal(signal domain.Signal) error
	Close() error
}
