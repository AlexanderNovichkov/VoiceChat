import copy
import dataclasses
import logging
import queue
import socket
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from typing import Optional

import pyaudio as pyaudio

import protocol
from gen import messages_pb2

logger = logging.getLogger(__name__)

CHUNK_SIZE = 1024
AUDIO_FORMAT = pyaudio.paInt16
CHANNELS = 1
RATE = 20000


@dataclasses.dataclass
class User:
    id: int
    name: str

    @staticmethod
    def from_protobuf(pb: messages_pb2.User):
        return User(id=pb.id, name=pb.name)


@dataclasses.dataclass
class Room:
    id: int
    users: Optional[list[User]]

    @staticmethod
    def from_protobuf(pb: messages_pb2.Room):
        return Room(
            id=pb.id,
            users=None if pb.users is None else [User.from_protobuf(u) for u in pb.users]
        )


@dataclasses.dataclass
class Status:
    rooms_ids: Optional[list[int]]
    room: Optional[Room]

    @staticmethod
    def from_protobuf(pb: messages_pb2.Status):
        return Status(
            rooms_ids=None if pb.rooms_ids is None else [id for id in pb.rooms_ids],
            room=None if pb.is_in_room is False else Room.from_protobuf(pb.room)
        )


class MultiplePeopleVoicePlayer:
    @dataclasses.dataclass
    class UserPlayerData:
        thread: threading.Thread
        queue: queue.Queue
        last_write_time_in_ms: float

    def __init__(self, p: pyaudio.PyAudio):
        self._p = p
        self._m = threading.Lock()
        self._user_id_to_data: dict[int, MultiplePeopleVoicePlayer.UserPlayerData] = dict()

    def write_user_data(self, user_id: int, data: bytes):
        with self._m:
            if user_id not in self._user_id_to_data:
                user_player_data = MultiplePeopleVoicePlayer.UserPlayerData(None, queue.Queue(), 0)
                user_player_data.t = threading.Thread(target=self._play_user_voice, args=(user_player_data,))
                user_player_data.t.start()
                self._user_id_to_data[user_id] = user_player_data
            self._user_id_to_data[user_id].queue.put(data)

    def get_speaking_users_ids(self) -> list[int]:
        users_ids = []
        with self._m:
            cur_time = time.time_ns() / 1_000_000
            for user_id, user_player_data in self._user_id_to_data.items():
                if (cur_time - user_player_data.last_write_time_in_ms) < CHUNK_SIZE / RATE * 1000 * 2:
                    users_ids.append(user_id)
        return users_ids

    def _play_user_voice(self, user_player_data: UserPlayerData):
        playing_stream = self._p.open(format=AUDIO_FORMAT, channels=CHANNELS, rate=RATE, output=True,
                                      frames_per_buffer=CHUNK_SIZE)
        try:
            while True:
                voice_data = user_player_data.queue.get()
                if voice_data is None:
                    break
                user_player_data.last_write_time_in_ms = time.time_ns() / 1_000_000
                playing_stream.write(voice_data)
        finally:
            playing_stream.stop_stream()
            playing_stream.close()

    def close(self):
        with self._m:
            for user_player_data in self._user_id_to_data.values():
                user_player_data.queue.put(None)
            for user_player_data in self._user_id_to_data.values():
                user_player_data.thread.join()


class Client:
    def __init__(self, server_ip: str, server_port: int, sing_in_token: str = None, sign_up_username: str = None):
        super().__init__()
        assert (sing_in_token is None) + (sign_up_username is None) == 1

        self.server_ip = server_ip
        self.server_port = server_port

        self._status: Optional[messages_pb2.Status] = None
        self._player: Optional[MultiplePeopleVoicePlayer] = None
        self._close = False
        self._is_muted = False

        try:
            self._s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self._s.connect((self.server_ip, self.server_port))
            self._f = self._s.makefile('rwb', buffering=0)

            if sign_up_username is not None:
                protocol.write_protobuf_message(messages_pb2.SignUpRequest(username=sign_up_username), self._f)
            else:
                raise NotImplementedError
            auth_response = protocol.read_transport_message(self._f).to_protobuf()
            if type(auth_response) is not messages_pb2.AuthorizationResponse:
                raise TypeError('Expected AuthorizationResponse')
            if not auth_response.ok:
                raise ValueError(f'AuthorizationResponse.ok = false: {auth_response.reason}')
            self.user_id: int = auth_response.user_id
            self.username: str = auth_response.username

            self._p: pyaudio.PyAudio = pyaudio.PyAudio()

            self._executor = ThreadPoolExecutor(5)
            self._executor.submit(self.send_voice_record_to_server)
            self._executor.submit(self.receive_server_data)
        except Exception:
            self.close()
            raise

        logger.debug(f'Connected to Server, username: {self.username}')

    def get_status(self) -> Status:
        return copy.deepcopy(self._status)

    def get_speaking_users_ids(self):
        return self._player.get_speaking_users_ids()

    def mute(self):
        self._is_muted = True

    def unmute(self):
        self._is_muted = False

    def close(self):
        self._close = True
        self._s.close()
        self._executor.shutdown(wait=True, cancel_futures=True)
        self._p.terminate()
        logger.debug(f'Closed, username: {self.username}')

    def join_room(self, room_id: int):
        message = messages_pb2.JoinRoomRequest(room_id=room_id)
        self._submit_message_using_thread_pool(message)

    def leave_room(self):
        message = messages_pb2.LeaveRoomRequest()
        self._submit_message_using_thread_pool(message)

    def create_room(self):
        message = messages_pb2.CreateRoomRequest()
        self._submit_message_using_thread_pool(message)

    def send_voice_record_to_server(self):
        recording_stream = self._p.open(format=AUDIO_FORMAT, channels=CHANNELS, rate=RATE, input=True,
                                        frames_per_buffer=CHUNK_SIZE)
        try:
            while not self._close:
                data = recording_stream.read(CHUNK_SIZE, exception_on_overflow=False)
                if not self._is_muted:
                    message = messages_pb2.SoundPacket(user_id=self.user_id, data=data)
                    protocol.write_protobuf_message(message, self._f)
        finally:
            logger.debug('send_voice_record_to_server - exit')
            recording_stream.close()
            self._s.close()

    def receive_server_data(self):
        self._player = MultiplePeopleVoicePlayer(self._p)
        try:
            while not self._close:
                message = protocol.read_protobuf_message(self._f)
                if type(message) == messages_pb2.Status:
                    self._status = Status.from_protobuf(message)
                elif type(message) == messages_pb2.SoundPacket:
                    self._player.write_user_data(message.user_id, message.data)
        finally:
            logger.debug(' receive_server_data - exit')
            self._player.close()
            self._s.close()

    def _submit_message_using_thread_pool(self, message):
        def submit_message():
            protocol.write_protobuf_message(message, self._f)

        self._executor.submit(submit_message)
