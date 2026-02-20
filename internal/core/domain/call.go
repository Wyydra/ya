package domain

type CallIntent string

const (
	IntentJoin CallIntent = "join_call" // i want to join call
	IntentNetwork  CallIntent = "network_map"  // how to join 
)

type CallNegotiation struct {
	UserID  UserID
	RoomID  RoomID
	Intent  CallIntent
	Payload []byte
}
