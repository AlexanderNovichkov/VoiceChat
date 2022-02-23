import logging
import socket

import protocol
from gen import messages_pb2

logger = logging.getLogger(__name__)


class Client:
    def __init__(self, server_ip: str, server_port: int, sing_in_token: str = None, sign_up_username: str = None):
        assert (sing_in_token is None) + (sign_up_username is None) == 1

        try:
            self.s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.s.connect((server_ip, server_port))
            self.f = self.s.makefile('rwb', buffering=0)
        except:
            logger.debug("Could not connect to server")
            raise

        if sign_up_username is not None:
            message = messages_pb2.SignUpRequest(username=sign_up_username)
            protocol.write_protobuf_message(message, self.f)
            auth_response = protocol.read_transport_message(self.f).to_protobuf()
            if type(auth_response) is not messages_pb2.AuthorizationResponse:
                raise TypeError("Expected AuthorizationResponse")


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
    #     print("Connected to Server")
    #
    #     # start threads
    #     receive_thread = threading.Thread(target=self.receive_server_data).start()
    #     self.send_data_to_server()
    #
    # def receive_server_data(self):
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
