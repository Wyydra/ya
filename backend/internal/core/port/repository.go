package port

import (
	"context"

	"github.com/Wyydra/ya/backend/internal/core/domain"
)

type MessageRepository interface {
	Save(ctx context.Context, msg domain.Message) error
}
