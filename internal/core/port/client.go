package port

import "github.com/Wyydra/ya/internal/core/domain"

type Client interface {
	ID() string
	Send(msg domain.Message) error
	Close() error
}

