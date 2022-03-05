package app

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"net"
	"server/gen"
	"server/protocol"
	"time"
)

type App struct {
	users    userPool
	sessions sessionPool
	rooms    roomPool
}

func (app *App) Run(port string) {
	fmt.Println("Launching server on port", port)

	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go app.handle(c)
	}
}

func (app *App) handle(c net.Conn) {
	fmt.Println("Handler started")
	s := &session{
		ToConnectionForwarder:   newToConnectionForwarder(c),
		FromConnectionForwarder: newFromConnectionForwarder(c),
		Connection:              c,
	}

	defer func() {
		if err := c.Close(); err != nil {
			fmt.Println("Connection.Close() error:", err)
		}
		if s.RoomInside != nil {
			s.RoomInside.removeUser(s.User)
		}
		if err := recover(); err != nil {
			fmt.Println("Recover from panic:", err)
		}
		fmt.Println("Client disconnected")
	}()

	app.authorizeUser(s)
	app.handleSession(s)
}

func (app *App) authorizeUser(s *session) {
	transportMessage := <-s.FromConnectionForwarder.Channel
	if transportMessage.Type != uint32(gen.MessageType_SIGN_UP_REQUEST) {
		panic("expected MessageType_SIGN_UP_REQUEST")
	}

	signUpRequest := &gen.SignUpRequest{}
	if err := proto.Unmarshal(transportMessage.Data, signUpRequest); err != nil {
		panic("proto.Unmarshal error:" + err.Error())
	}

	s.User = app.users.addUser(&user{Name: signUpRequest.Username})

	authorizationResponse := &gen.AuthorizationResponse{Ok: true, UserId: s.User.Id, Username: s.User.Name}
	authorizationResponseTransportMessage, err := protocol.NewTransportMessageFromProtobuf(
		gen.MessageType_AUTHORIZATION_RESPONSE, authorizationResponse,
	)
	if err != nil {
		panic("protocol.NewTransportMessageFromProtobuf error in authorization:" + err.Error())
	}
	s.ToConnectionForwarder.Channel <- &authorizationResponseTransportMessage
	fmt.Println("User authorized!", s.User)
}

func (app *App) handleSession(session *session) {
	ticker := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case <-ticker.C:
			app.sendStatusToUser(session)
		case transportMessage, ok := <-session.FromConnectionForwarder.Channel:
			if !ok {
				return
			}
			app.handleMessageFromClient(session, transportMessage)
		}
	}
}

func (app *App) handleMessageFromClient(session *session, transportMessage *protocol.TransportMessage) {
	switch gen.MessageType(transportMessage.Type) {
	case gen.MessageType_CREATE_ROOM_REQUEST:
		message := &gen.CreateRoomRequest{}
		if err := proto.Unmarshal(transportMessage.Data, message); err != nil {
			break
		}
		room := newRoom()
		app.rooms.createNewRoom(room)
		responseMessage := gen.CreateRoomResponse{RoomId: room.Id}
		responseTransportMessage, _ := protocol.NewTransportMessageFromProtobuf(
			gen.MessageType_CREATE_ROOM_RESPONSE, &responseMessage,
		)
		session.ToConnectionForwarder.Channel <- &responseTransportMessage

	case gen.MessageType_JOIN_ROOM_REQUEST:
		message := &gen.JoinRoomRequest{}
		if err := proto.Unmarshal(transportMessage.Data, message); err != nil {
			break
		}
		room, ok := app.rooms.getRoom(uint32(message.RoomId))
		if !ok {
			break
		}
		if session.RoomInside != nil {
			session.RoomInside.removeUser(session.User)
			session.RoomInside = nil
		}
		room.addUser(session.User, session.ToConnectionForwarder.Channel)
		session.RoomInside = room

	case gen.MessageType_LEAVE_ROOM_REQUEST:
		message := &gen.LeaveRoomRequest{}
		if err := proto.Unmarshal(transportMessage.Data, message); err != nil {
			break
		}
		if session.RoomInside != nil {
			session.RoomInside.removeUser(session.User)
			session.RoomInside = nil
		}

	case gen.MessageType_SOUND_PACKET:
		if session.RoomInside != nil {
			session.RoomInside.getInputChannel() <- transportMessage
		}
	}
}

func (app *App) sendStatusToUser(session *session) {
	status := gen.Status{
		RoomsIds: app.rooms.getRoomsIds(),
		IsInRoom: false,
		Room:     nil,
	}
	if session.RoomInside != nil {
		status.IsInRoom = true
		status.Room = session.RoomInside.toProtobufMessage()
	}

	transportMessage, err := protocol.NewTransportMessageFromProtobuf(gen.MessageType_STATUS, &status)
	if err != nil {
		panic(err)
	}

	session.ToConnectionForwarder.Channel <- &transportMessage
}
