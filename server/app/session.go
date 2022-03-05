package app

import (
	"net"
	"sync"
)

type session struct {
	User                    *user
	ToConnectionForwarder   toConnectionForwarder
	FromConnectionForwarder fromConnectionForwarder
	Connection              net.Conn
	RoomInside              *room
}

type sessionPool struct {
	m               sync.Mutex
	userIdToSession map[uint32]*session
}

func (sessionPool *sessionPool) updateClientSession(usedId uint32, session *session) {
	sessionPool.m.Lock()
	defer sessionPool.m.Unlock()
	oldSession, ok := sessionPool.userIdToSession[usedId]
	if ok {
		_ = oldSession.Connection.Close()
	}
	sessionPool.userIdToSession[usedId] = session
}
