import logging
import socket
import threading
import typing
from concurrent.futures import Future, ThreadPoolExecutor
from typing import Any, Optional

import protocol
from gen import messages_pb2

logger = logging.getLogger(__name__)


class Client(threading.Thread):
    def __init__(self, server_ip: str, server_port: int, sing_in_token: str = None, sign_up_username: str = None):
        super().__init__()
        assert (sing_in_token is None) + (sign_up_username is None) == 1
        self.server_ip = server_ip
        self.server_port = server_port
        self.status: Optional[messages_pb2.Status] = None
        self.user_id: Optional[int] = None
        self.username: Optional[int] = None
        self._s: Optional[socket] = None
        self._f = None
        self._auth_request_message: Any[messages_pb2.SignUpRequest, messages_pb2.SignInRequest]
        self._executor: Optional[ThreadPoolExecutor] = None
        self._initialized: threading.Event = threading.Event()
        self._shutdown: threading.Event = threading.Event()

        if sign_up_username is not None:
            self._auth_request_message = messages_pb2.SignUpRequest(username=sign_up_username)
        else:
            raise NotImplementedError

    def run(self):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as self._s:
            self._s.connect((self.server_ip, self.server_port))
            self._f = self._s.makefile('rwb', buffering=0)

            protocol.write_protobuf_message(self._auth_request_message, self._f)
            auth_response = protocol.read_transport_message(self._f).to_protobuf()
            if type(auth_response) is not messages_pb2.AuthorizationResponse:
                raise TypeError('Expected AuthorizationResponse')
            if not auth_response.ok:
                raise ValueError(f'AuthorizationResponse.ok = false: {auth_response.reason}')
            self.user_id = auth_response.user_id
            self.username = auth_response.username

            self._executor = ThreadPoolExecutor()

            f1 = self._executor.submit(self.send_voice_record_to_server)
            f2 = self._executor.submit(self.receive_server_data)
            f1.add_done_callback(lambda future: self.shutdown())
            f2.add_done_callback(lambda future: self.shutdown())

            self._initialized.set()

            self._shutdown.wait()
            self._s.close()
            self._executor.shutdown()

    #     chunk_size = 1024  # 512
    #     audio_format = pyaudio.paInt16
    #     channels = 1
    #     rate = 20000
    #
    #     # initialise microphone recording
    #     self.p = pyaudio.PyAudio()
    #     self.playing_stream = self.p.open(format=audio_format, channels=channels, rate=rate, output=True,
    #                                       frames_per_buffer=chunk_size)
    #     self.recording_stream = self.p.open(format=audio_format, channels=channels, rate=rate, input=True,
    #                                         frames_per_buffer=chunk_size)
    #
    #     print('Connected to Server')
    #
    #     # start threads
    #     receive_thread = threading.Thread(target=self.receive_server_data).start()
    #     self.send_data_to_server()
    #
    # def receive_data_from_server(self):
    #     while True:
    #         try:
    #             data = self.s.recv(1024)
    #             self.playing_stream.write(data)
    #         except:
    #             pass
    #
    # def send_data_to_server(self):
    #     while True:
    #         try:
    #             data = self.recording_stream.read(1024)
    #             self.s.sendall(data)
    #         except:
    #             pass

    def wait_for_initialization(self):
        self._initialized.wait()

    def shutdown(self):
        print("Shutdown")
        self._shutdown.set()

    def join_room(self, room_id: int):
        message = messages_pb2.JoinRoomRequest(room_id=room_id)

        def submit():
            protocol.write_protobuf_message(message, self._f)

        self._run_in_executor(submit)

    def create_room(self):
        message = messages_pb2.CreateRoomRequest()

        def submit():
            protocol.write_protobuf_message(message, self._f)

        self._run_in_executor(submit)

    def send_voice_record_to_server(self):
        while True:
            data = input()
            message = messages_pb2.SoundPacket(user_id=self.user_id, data=data.encode())
            protocol.write_protobuf_message(message, self._f)

    def receive_server_data(self):
        while True:
            message = protocol.read_protobuf_message(self._f)
            if type(message) == messages_pb2.Status:
                self.status = message
            elif type(message) == messages_pb2.SoundPacket:
                print(message.data.decode())

    def _run_in_executor(self, f: typing.Callable, *args, **kwargs):
        def callback(future: Future):
            try:
                future.result()
            except:
                self.shutdown()

        self._executor.submit(f, *args, **kwargs).add_done_callback(callback)
