package app

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"net"
	"server/gen"
	"server/protocol"
)

type App struct {
	users    UserPool
	sessions SessionPool
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
		fmt.Println("Client disconnected")
		if err := c.Close(); err != nil {
			fmt.Println("c.Close() error:", err)
		}
		close(s.FromClient)
		close(s.ToClient)
	}()

	finished := make(chan bool)
	go func() {
		forwardMessagesFromConnectionToChannel(s.c, s.FromClient)
		finished <- true
	}()
	go func() {
		forwardMessagesFromChannelToConnection(s.ToClient, s.c)
		finished <- true
	}()

	app.authorizeUser(s)

	<-finished
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

	s.User = app.users.AddUser(User{Name: signUpRequest.Username})

	authorizationResponse := &gen.AuthorizationResponse{Ok: true, UserId: s.User.Id}
	authorizationResponseTransportMessage, err := protocol.NewTransportMessageFromProtobuf(
		gen.MessageType_AUTHORIZATION_RESPONSE, authorizationResponse,
	)
	if err != nil {
		panic("protocol.NewTransportMessageFromProtobuf error in authorization:" + err.Error())
	}
	s.ToClient <- &authorizationResponseTransportMessage
	fmt.Println("User authorized!", s.User)
}

func (app *App) handleMessagesFromUser(c net.Conn, user User) {
	for {
		transportMessage, err := protocol.ReadTransportMessage(c)
		if err != nil {
			return
		}
		switch gen.MessageType(transportMessage.Type) {
		case gen.MessageType_SOUND_PACKET:

		}
	}
}
