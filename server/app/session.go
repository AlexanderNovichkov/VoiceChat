package app

import (
	"net"
	"server/protocol"
	"sync"
)

type Session struct {
	User       User
	ToClient   chan *protocol.TransportMessage
	FromClient chan *protocol.TransportMessage
	c          net.Conn
}

type SessionPool struct {
	m               sync.Mutex
	userIdToSession map[uint32]Session
}

func (sessionPool *SessionPool) UpdateClientSession(usedId uint32, session *Session) {
	sessionPool.m.Lock()
	defer sessionPool.m.Unlock()
	old_session, ok := sessionPool.userIdToSession[usedId]
	if ok {
		_ = old_session.c.Close()
	}
	sessionPool.userIdToSession[usedId] = *session
}
