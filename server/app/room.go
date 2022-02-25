package app

import (
	"server/gen"
	"server/protocol"
	"sync"
)

type Room struct {
	UserToChannel sync.Map
	InputChannel  chan *protocol.TransportMessage
	Id            uint32
}

func NewRoom() *Room {
	return &Room{
		InputChannel: make(chan *protocol.TransportMessage),
	}
}

func (room *Room) AddUserToBroadcast(user *User, channel chan<- *protocol.TransportMessage) {
	room.UserToChannel.Store(user, channel)
}

func (room *Room) RemoveUserFromBroadcast(user *User) {
	room.UserToChannel.Delete(user)
}

func (room *Room) StreamDataToUsers() {
	for message := range room.InputChannel {
		room.UserToChannel.Range(func(user, channel interface{}) bool {
			defer func() {
				recover()
			}()
			select {
			case channel.(chan<- *protocol.TransportMessage) <- message:
			default:
			}
			return true
		})
	}
}

func (room *Room) ToProtobufMessage() *gen.Room {
	message := gen.Room{Id: room.Id}
	room.UserToChannel.Range(func(user, channel interface{}) bool {
		message.Users = append(message.Users, &gen.User{
			Id:   user.(*User).Id,
			Name: user.(*User).Name,
		})
		return true
	})
	return &message
}

func (room *Room) GetInputChannel() chan<- *protocol.TransportMessage {
	return room.InputChannel
}

type roomPool struct {
	m        sync.Mutex
	idToRoom []*Room
}

func (roomPool *roomPool) GetRoom(id uint32) (*Room, bool) {
	roomPool.m.Lock()
	defer roomPool.m.Unlock()
	if id < 0 || id >= uint32(len(roomPool.idToRoom)) {
		return nil, false
	}
	return roomPool.idToRoom[id], true
}

func (roomPool *roomPool) CreateNewRoom(room *Room) {
	roomPool.m.Lock()
	defer roomPool.m.Unlock()
	room.Id = uint32(len(roomPool.idToRoom))
	roomPool.idToRoom = append(roomPool.idToRoom, room)
	go room.StreamDataToUsers()
}

func (roomPool *roomPool) GetRoomsIds() (ids []uint32) {
	for _, room := range roomPool.idToRoom {
		ids = append(ids, room.Id)
	}
	return
}
