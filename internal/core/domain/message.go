package domain

import (
	"errors"
)

type Message struct {
	ID       MessageID
	RoomID   RoomID
	SenderID UserID
	Content  string
	// CreatedAt time.Time
}

func NewMessage(senderID UserID, roomID RoomID, content string) (*Message, error) {
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}
	return &Message{
		ID:       NewMessageID(),
		RoomID:   roomID,
		SenderID: senderID,
		Content:  content,
	}, nil
}
