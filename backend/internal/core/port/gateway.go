package port

import (
	"context"

	"github.com/Wyydra/ya/backend/internal/core/domain"
)

type RealTimeGateway interface {
	BroadcastMessage(ctx context.Context, msg domain.Message) error
	SendSignal(ctx context.Context, userID domain.UserID, signal domain.Signal) error
	NotifyUserJoined(ctx context.Context, roomID domain.RoomID, userID domain.UserID) error
}
