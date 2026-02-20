package domain

import (
	"github.com/google/uuid"
)

type UserID uuid.UUID
type RoomID uuid.UUID

func NewUserID() UserID {
	return UserID(uuid.New())
}

func NewRoomID() RoomID {
	return RoomID(uuid.New())
}

func (id UserID) String() string {
	return uuid.UUID(id).String()
}

type MessageID uuid.UUID

func NewMessageID() MessageID {
	return MessageID(uuid.New())
}

func (id MessageID) String() string {
	return uuid.UUID(id).String()
}
