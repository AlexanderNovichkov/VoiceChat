package app

import (
	"net"
	"server/protocol"
)

func forwardMessagesFromConnectionToChannel(c net.Conn, channel chan<- *protocol.TransportMessage) {
	defer func() {
		recover()
	}()
	for {
		transportMessage, err := protocol.ReadTransportMessage(c)
		if err != nil {
			break
		}
		channel <- transportMessage
	}
}

func forwardMessagesFromChannelToConnection(channel <-chan *protocol.TransportMessage, c net.Conn) {
	defer func() {
		recover()
	}()
	for {
		transportMessage := <-channel
		if err := protocol.WriteTransportMessage(transportMessage, c); err != nil {
			break
		}
	}
}
