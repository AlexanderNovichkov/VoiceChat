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
	users    UserPool
	sessions SessionPool
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
	s := &Session{
		ToClient:   make(chan *protocol.TransportMessage),
		FromClient: make(chan *protocol.TransportMessage),
		c:          c,
	}

	defer func() {
		if err := c.Close(); err != nil {
			fmt.Println("c.Close() error:", err)
		}
		close(s.FromClient)
		close(s.ToClient)
		if s.roomInside != nil {
			s.roomInside.RemoveUserFromBroadcast(s.User)
		}
		if err := recover(); err != nil {
			fmt.Println("Recover from panic:", err)
		}
		fmt.Println("Client disconnected")
	}()

	go forwardMessagesFromConnectionToChannel(s.c, s.FromClient)
	go forwardMessagesFromChannelToConnection(s.ToClient, s.c)

	app.authorizeUser(s)
	app.handleSession(s)
}

func (app *App) authorizeUser(s *Session) {
	transportMessage := <-s.FromClient
	if transportMessage.Type != uint32(gen.MessageType_SIGN_UP_REQUEST) {
		panic("expected MessageType_SIGN_UP_REQUEST")
	}

	signUpRequest := &gen.SignUpRequest{}
	if err := proto.Unmarshal(transportMessage.Data, signUpRequest); err != nil {
		panic("proto.Unmarshal error:" + err.Error())
	}

	s.User = app.users.AddUser(&User{Name: signUpRequest.Username})

	authorizationResponse := &gen.AuthorizationResponse{Ok: true, UserId: s.User.Id, Username: s.User.Name}
	authorizationResponseTransportMessage, err := protocol.NewTransportMessageFromProtobuf(
		gen.MessageType_AUTHORIZATION_RESPONSE, authorizationResponse,
	)
	if err != nil {
		panic("protocol.NewTransportMessageFromProtobuf error in authorization:" + err.Error())
	}
	s.ToClient <- &authorizationResponseTransportMessage
	fmt.Println("User authorized!", s.User)
}

func (app *App) handleSession(session *Session) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for i := 1; ; i++ {
		select {
		case <-ticker.C:
			fmt.Println("SendStatus", i)
			app.SendStatusToUser(session)
		case transportMessage := <-session.FromClient:
			fmt.Println("handleMessageFromClient", i)
			app.handleMessageFromClient(session, transportMessage)
		}
	}
}

func (app *App) handleMessageFromClient(session *Session, transportMessage *protocol.TransportMessage) {
	switch gen.MessageType(transportMessage.Type) {
	case gen.MessageType_CREATE_ROOM_REQUEST:
		message := &gen.CreateRoomRequest{}
		if err := proto.Unmarshal(transportMessage.Data, message); err != nil {
			break
		}
		room := NewRoom()
		app.rooms.CreateNewRoom(room)
		responseMessage := gen.CreateRoomResponse{RoomId: room.Id}
		responseTransportMessage, _ := protocol.NewTransportMessageFromProtobuf(
			gen.MessageType_CREATE_ROOM_RESPONSE, &responseMessage,
		)
		session.ToClient <- &responseTransportMessage

	case gen.MessageType_JOIN_ROOM_REQUEST:
		message := &gen.JoinRoomRequest{}
		if err := proto.Unmarshal(transportMessage.Data, message); err != nil {
			break
		}
		room, ok := app.rooms.GetRoom(uint32(message.RoomId))
		if !ok {
			break
		}
		if session.roomInside != nil {
			session.roomInside.RemoveUserFromBroadcast(session.User)
			session.roomInside = nil
		}
		room.AddUserToBroadcast(session.User, session.ToClient)
		session.roomInside = room

	case gen.MessageType_LEAVE_ROOM_REQUEST:
		message := &gen.LeaveRoomRequest{}
		if err := proto.Unmarshal(transportMessage.Data, message); err != nil {
			break
		}
		if session.roomInside != nil {
			session.roomInside.RemoveUserFromBroadcast(session.User)
			session.roomInside = nil
		}

	case gen.MessageType_SOUND_PACKET:
		if session.roomInside != nil {
			session.roomInside.GetInputChannel() <- transportMessage
		}
	}
}

func (app *App) SendStatusToUser(session *Session) {
	status := gen.Status{
		RoomsIds: app.rooms.GetRoomsIds(),
		IsInRoom: false,
		Room:     nil,
	}
	if session.roomInside != nil {
		status.IsInRoom = true
		status.Room = session.roomInside.ToProtobufMessage()
	}

	transportMessage, err := protocol.NewTransportMessageFromProtobuf(gen.MessageType_STATUS, &status)
	if err != nil {
		panic(err)
	}

	session.ToClient <- &transportMessage
}

func has(users []*gen.User, userId uint32) bool {
	for _, user := range users {
		if user.Id == userId {
			return true
		}
	}
	return false
}
