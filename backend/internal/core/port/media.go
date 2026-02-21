package port

import "github.com/Wyydra/ya/backend/internal/core/domain"

type MediaEngine interface {
	AddPeer(sessionID domain.SessionID, userID domain.UserID) (offer domain.Signal, err error)
	HandleSignal(sessionID domain.SessionID, userID domain.UserID, signal domain.Signal) error
	RemovePeer(sessionID domain.SessionID,userID domain.UserID)
	SetSignalCallback(cb func(sessionID domain.SessionID, userID domain.UserID, signal domain.Signal)) //TODO: investigate if its needed
}
