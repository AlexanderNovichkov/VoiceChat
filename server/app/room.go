package app

import "server/protocol"

type Room struct {
	UsersChannels map[User]chan protocol.TransportMessage
	RoomChannel   chan protocol.TransportMessage
}

func (room *Room) AddUser(user User, channel chan<- protocol.TransportMessage) {
}

func (room *Room) StreamDataToUsers() {
}
