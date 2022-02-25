import logging
import socket
import time

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
    client.start()
    client.wait_for_initialization()
    client.create_room()
    client.join_room(0)
    while client.is_alive():
        # print(client.status)
        time.sleep(4)

    print("Dead")



if __name__ == '__main__':
    main()
