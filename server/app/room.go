package app

import (
	"server/gen"
	"server/protocol"
	"sync"
)

type room struct {
	userToChannel sync.Map
	InputChannel  chan *protocol.TransportMessage
	Id            uint32
}

func newRoom() *room {
	return &room{
		InputChannel: make(chan *protocol.TransportMessage),
	}
}

func (room *room) addUser(user *user, channel chan<- *protocol.TransportMessage) {
	room.userToChannel.Store(user, channel)
}

func (room *room) removeUser(user *user) {
	room.userToChannel.Delete(user)
}

func (room *room) streamDataToUsers() {
	for message := range room.InputChannel {
		room.userToChannel.Range(func(user, channel interface{}) bool {
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

func (room *room) toProtobufMessage() *gen.Room {
	message := gen.Room{Id: room.Id}
	room.userToChannel.Range(func(u, channel interface{}) bool {
		message.Users = append(message.Users, &gen.User{
			Id:   u.(*user).Id,
			Name: u.(*user).Name,
		})
		return true
	})
	return &message
}

func (room *room) getInputChannel() chan<- *protocol.TransportMessage {
	return room.InputChannel
}

type roomPool struct {
	m        sync.Mutex
	idToRoom []*room
}

func (roomPool *roomPool) getRoom(id uint32) (*room, bool) {
	roomPool.m.Lock()
	defer roomPool.m.Unlock()
	if id < 0 || id >= uint32(len(roomPool.idToRoom)) {
		return nil, false
	}
	return roomPool.idToRoom[id], true
}

func (roomPool *roomPool) createNewRoom(room *room) {
	roomPool.m.Lock()
	defer roomPool.m.Unlock()
	room.Id = uint32(len(roomPool.idToRoom))
	roomPool.idToRoom = append(roomPool.idToRoom, room)
	go room.streamDataToUsers()
}

func (roomPool *roomPool) getRoomsIds() (ids []uint32) {
	for _, room := range roomPool.idToRoom {
		ids = append(ids, room.Id)
	}
	return
}
