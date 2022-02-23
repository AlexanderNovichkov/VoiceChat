import logging
import socket

from client import Client
from gen import messages_pb2
import protocol
# import pyaudio

logger = logging.getLogger(__name__)


def main():
    logging.basicConfig(level=logging.DEBUG)
    # ip = input('Enter IP address of server --> ')
    # port = int(input('Enter target port of server --> '))
    ip, port = 'localhost', 8081
    client = Client(ip, port, sign_up_username="Alex")


if __name__ == '__main__':
    main()
