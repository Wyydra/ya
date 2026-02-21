package memory

import (
	"context"
	"sync"
	"github.com/Wyydra/ya/backend/internal/core/domain"
)

type MessageRepository struct {
	mu       sync.Mutex
	messages []domain.Message
}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{
		messages: make([]domain.Message, 0),
	}
}

func (r *MessageRepository) Save(ctx context.Context, msg domain.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, msg)
	return nil
}
