package port

import (
	"context"

	"github.com/Wyydra/ya/internal/core/domain"
)

type RealTimeGateway interface {
	BroadcastMessage(ctx context.Context, msg domain.Message) error
	SendCallSignal(ctx context.Context, userID domain.UserID, negotiation domain.CallNegotiation) error
	NotifyUserJoined(ctx context.Context, roomID domain.RoomID, userID domain.UserID) error
}
