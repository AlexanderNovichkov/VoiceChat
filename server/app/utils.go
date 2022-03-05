package app

import (
	"net"
	"server/protocol"
)

type fromConnectionForwarder struct {
	Channel    <-chan *protocol.TransportMessage
	connection net.Conn
}

type toConnectionForwarder struct {
	Channel    chan<- *protocol.TransportMessage
	connection net.Conn
}

func newFromConnectionForwarder(connection net.Conn) fromConnectionForwarder {
	channel := make(chan *protocol.TransportMessage)
	go func() {
		defer func() {
			close(channel)
			recover()
		}()
		for {
			transportMessage, err := protocol.ReadTransportMessage(connection)
			if err != nil {
				break
			}
			channel <- transportMessage
		}
	}()
	return fromConnectionForwarder{Channel: channel, connection: connection}
}

func newToConnectionForwarder(connection net.Conn) toConnectionForwarder {
	channel := make(chan *protocol.TransportMessage)
	go func() {
		defer func() {
			close(channel)
			recover()
		}()
		for transportMessage := range channel {
			if err := protocol.WriteTransportMessage(transportMessage, connection); err != nil {
				break
			}
		}
	}()
	return toConnectionForwarder{Channel: channel, connection: connection}
}
