package protocol

import (
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"io"
	"server/gen"
)

type TransportMessage struct {
	Type uint32
	Data []byte
}

func ReadTransportMessage(reader io.Reader) (*TransportMessage, error) {
	typeAndLength := make([]byte, 8)
	if _, err := io.ReadFull(reader, typeAndLength); err != nil {
		return nil, err
	}

	messageType := binary.BigEndian.Uint32(typeAndLength[:4])
	messageLength := binary.BigEndian.Uint32(typeAndLength[4:])
	messageData := make([]byte, messageLength)
	if _, err := io.ReadFull(reader, messageData); err != nil {
		return nil, err
	}

	return &TransportMessage{messageType, messageData}, nil
}

func WriteTransportMessage(transportMessage *TransportMessage, writer io.Writer) error {
	typeAndLength := make([]byte, 8)
	binary.BigEndian.PutUint32(typeAndLength[:4], transportMessage.Type)
	binary.BigEndian.PutUint32(typeAndLength[4:], uint32(len(transportMessage.Data)))
	if err := writeFull(writer, typeAndLength); err != nil {
		return err
	}
	if err := writeFull(writer, transportMessage.Data); err != nil {
		return err
	}
	return nil
}

func NewTransportMessageFromProtobuf(messageType gen.MessageType, message proto.Message) (TransportMessage, error) {
	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return TransportMessage{}, err
	}
	return TransportMessage{uint32(messageType), encodedMessage}, nil
}

func writeFull(writer io.Writer, buf []byte) error {
	done := 0
	for done < len(buf) {
		n, err := writer.Write(buf[done:])
		if err != nil {
			return err
		}
		done += n
	}
	return nil
}
