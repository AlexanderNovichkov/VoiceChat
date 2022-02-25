package app

import (
	"net"
	"server/protocol"
	"sync"
)

type Session struct {
	User       *User
	ToClient   chan *protocol.TransportMessage
	FromClient chan *protocol.TransportMessage
	c          net.Conn
	roomInside *Room
}

type SessionPool struct {
	m               sync.Mutex
	userIdToSession map[uint32]*Session
}

func (sessionPool *SessionPool) UpdateClientSession(usedId uint32, session *Session) {
	sessionPool.m.Lock()
	defer sessionPool.m.Unlock()
	oldSession, ok := sessionPool.userIdToSession[usedId]
	if ok {
		_ = oldSession.c.Close()
	}
	sessionPool.userIdToSession[usedId] = session
}
