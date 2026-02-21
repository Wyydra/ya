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

// TODO: remove this
func NewRoomIDFromString(s string) (RoomID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return RoomID{}, err
	}
	return RoomID(id), nil
}

func (id UserID) String() string {
	return uuid.UUID(id).String()
}

func (id RoomID) String() string {
	return uuid.UUID(id).String()
}

type MessageID uuid.UUID

func NewMessageID() MessageID {
	return MessageID(uuid.New())
}

func (id MessageID) String() string {
	return uuid.UUID(id).String()
}
