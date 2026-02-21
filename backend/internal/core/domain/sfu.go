package domain

import "github.com/google/uuid"

type SessionID string

func NewSessionID() SessionID  {
	return SessionID(uuid.New().String())
}

func (s SessionID) String() string {
	return string(s)
}

type SignalType string

const (
	SignalOffer SignalType = "offer"
	SignalAnswer SignalType = "answer"
	SignalCandidate SignalType = "candidate"
)

type Signal struct {
	Type SignalType
	Payload string
}

func NewSignal(t SignalType, payload string) Signal {
	return Signal{
		Type: t,
		Payload: payload,
	}
}
