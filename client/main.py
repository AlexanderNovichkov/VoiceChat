import logging
import time

import pyaudio

# import pyaudio

logger = logging.getLogger(__name__)


def main():
    logging.basicConfig(level=logging.DEBUG, format='%(asctime)s %(funcName)20s() %(levelname)s %(message)s')
    # # ip = input('Enter IP address of server --> ')
    # # port = int(input('Enter target port of server --> '))
    # ip, port = 'localhost', 8081
    # client = Client(ip, port, sign_up_username="Alex")
    # client.create_room()
    # client.join_room(0)
    # time.sleep(10)
    # client.close()
    # time.sleep(1000)

    CHUNK_SIZE = 1024
    AUDIO_FORMAT = pyaudio.paInt16
    CHANNELS = 1
    RATE = 20000

    p: pyaudio.PyAudio = pyaudio.PyAudio()

    recording_stream = p.open(format=AUDIO_FORMAT, channels=CHANNELS, rate=RATE,
                              input=True,
                              frames_per_buffer=CHUNK_SIZE)

    playing_stream = p.open(format=AUDIO_FORMAT, channels=CHANNELS, rate=RATE,
                            output=True,
                            frames_per_buffer=CHUNK_SIZE * 1)

    print(CHUNK_SIZE / RATE)

    # time.sleep(0.5)

    for i in range(0, 100):
        start = time.time()
        r = recording_stream.read(CHUNK_SIZE, exception_on_overflow=False)
        playing_stream.write(r)
        print("{:.5f}".format(time.time() - start))



if __name__ == '__main__':
    main()
