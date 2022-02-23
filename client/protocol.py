import dataclasses
import struct
from typing import BinaryIO

from gen import messages_pb2

MESSAGE_TYPE_TO_PB_CLASS = {
    messages_pb2.SIGN_IN_REQUEST: messages_pb2.SignInRequest,
    messages_pb2.SIGN_UP_REQUEST: messages_pb2.SignUpRequest,
    messages_pb2.AUTHORIZATION_RESPONSE: messages_pb2.AuthorizationResponse,
    messages_pb2.JOIN_ROOM_REQUEST: messages_pb2.JoinRoomRequest,
    messages_pb2.SOUND_PACKET: messages_pb2.SoundPacket,
}

BP_CLASS_TO_MESSAGE_TYPE = {v: k for k, v in MESSAGE_TYPE_TO_PB_CLASS.items()}


@dataclasses.dataclass
class TransportMessage:
    message_type: int
    message_data: bytes

    def to_protobuf(self):
        if self.message_type not in MESSAGE_TYPE_TO_PB_CLASS:
            raise ValueError(f'Message type = {self.message_type} is not correct')
        pb_message = MESSAGE_TYPE_TO_PB_CLASS[self.message_type]()
        pb_message.ParseFromString(self.message_data)
        return pb_message

    @staticmethod
    def from_protobuf(pb_obj):
        return TransportMessage(BP_CLASS_TO_MESSAGE_TYPE[type(pb_obj)], pb_obj.SerializeToString())


def read_transport_message(f: BinaryIO):
    type_and_length: bytes = f.read(8)
    message_type, message_length = struct.unpack(">LL", type_and_length)
    message_data = f.read(message_length)
    return TransportMessage(message_type, message_data)


def write_transport_message(transport_message: TransportMessage, f: BinaryIO):
    type_and_length = struct.pack(">LL", transport_message.message_type, len(transport_message.message_data))
    f.write(type_and_length)
    f.write(transport_message.message_data)


def write_protobuf_message(pb_message, f: BinaryIO):
    write_transport_message(TransportMessage.from_protobuf(pb_message), f)


def read_protobuf_message(f: BinaryIO):
    return read_transport_message(f).to_protobuf()
