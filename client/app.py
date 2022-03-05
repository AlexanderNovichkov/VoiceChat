import logging
import sys

from PySide6.QtCore import QTimer, Qt
from PySide6.QtGui import QCloseEvent, QColor, QIntValidator
from PySide6.QtWidgets import (QApplication, QHBoxLayout, QLabel, QLineEdit, QListWidget, QListWidgetItem, QMainWindow,
                               QPushButton, QVBoxLayout, QWidget)

from client import Client, User

logger = logging.getLogger(__name__)


class QListWidgetItemUser(QListWidgetItem):
    def __init__(self, user: User):
        super().__init__(user.name)
        self.user = user


class MainWindow(QMainWindow):

    def __init__(self, parent: QWidget, client: Client):
        super().__init__()
        self.parent = parent
        self.client = client
        self.setWindowTitle("Voice Chat")

        self.setAttribute(Qt.WA_DeleteOnClose, True)

        self.rooms = QListWidget()
        self.rooms.itemDoubleClicked.connect(self.join_room)
        self.join_room_button = QPushButton("Join")
        self.join_room_button.clicked.connect(self.join_room)
        self.create_room_button = QPushButton("Create")
        self.create_room_button.clicked.connect(self.create_room)
        rooms_layout = QVBoxLayout()
        rooms_label = QLabel("Rooms")
        rooms_label.setAlignment(Qt.AlignCenter)
        rooms_layout.addWidget(rooms_label)
        rooms_layout.addWidget(self.rooms)
        rooms_layout.addWidget(self.join_room_button)
        rooms_layout.addWidget(self.create_room_button)

        self.current_room_label = QLabel("Out of room")
        self.current_room_label.setAlignment(Qt.AlignCenter)
        self.current_room_users = QListWidget()
        self.current_room_mute_unmute_button = QPushButton("Mute")
        self.current_room_mute_unmute_button.clicked.connect(self.mute_unmute)
        self.current_room_leave_button = QPushButton("Leave room")
        self.current_room_leave_button.clicked.connect(self.leave_room)
        user_room_layout = QVBoxLayout()
        user_room_layout.addWidget(self.current_room_label)
        user_room_layout.addWidget(self.current_room_users)
        user_room_layout.addWidget(self.current_room_mute_unmute_button)
        user_room_layout.addWidget(self.current_room_leave_button)

        layout = QHBoxLayout()
        layout.addLayout(rooms_layout, 2)
        layout.addLayout(user_room_layout, 3)

        self.timer = QTimer(self)
        self.timer.setInterval(500)
        self.timer.timeout.connect(self.update_window_with_new_data)
        self.timer.start()

        widget = QWidget()
        widget.setLayout(layout)
        self.setCentralWidget(widget)

    def closeEvent(self, event: QCloseEvent) -> None:
        self.client.close()
        event.accept()
        self.parent.show()

    def create_room(self):
        self.client.create_room()

    def join_room(self):
        item = self.rooms.currentItem()
        if item is not None:
            self.client.join_room(int(item.text()))

    def mute_unmute(self, event):
        if self.current_room_mute_unmute_button.text() == "Mute":
            self.client.mute()
            self.current_room_mute_unmute_button.setText("Unmute")
        else:
            self.client.unmute()
            self.current_room_mute_unmute_button.setText("Mute")

    def leave_room(self):
        self.client.leave_room()
        self.rooms.currentItem().setSelected(False)

    def update_window_with_new_data(self):
        status = self.client.get_status()
        if status is None:
            return

        for i in range(self.rooms.count(), len(status.rooms_ids)):
            self.rooms.addItem(str(status.rooms_ids[i]))

        self.current_room_users.clear()
        if status.room is None:
            self.current_room_label.setText("Out of room")
            self.current_room_leave_button.setDisabled(True)
            self.current_room_mute_unmute_button.setDisabled(True)
        else:
            speaking_users_ids = self.client.get_speaking_users_ids()
            self.current_room_label.setText(f"Current room: {status.room.id}")
            self.current_room_leave_button.setDisabled(False)
            self.current_room_mute_unmute_button.setDisabled(False)
            for user in status.room.users:
                item = QListWidgetItemUser(user)
                if user.id in speaking_users_ids:
                    item.setBackground(QColor('green'))
                self.current_room_users.addItem(item)
            self.current_room_users.sortItems()


class ServerConnectionWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        self.w = None

        self.setWindowTitle("Voice Chat")

        self.ip = QLineEdit()
        self.ip.setPlaceholderText("Server IP")

        self.port = QLineEdit()
        self.port.setPlaceholderText("Server port")
        self.port.setValidator(QIntValidator())

        self.username = QLineEdit()
        self.username.setPlaceholderText("Username")

        self.connect_button = QPushButton("Connect")
        self.connect_button.clicked.connect(self.connect_to_server)
        self.setCentralWidget(self.connect_button)

        self.status = QLabel()
        self.status.hide()

        layout = QVBoxLayout()
        layout.addWidget(self.ip)
        layout.addWidget(self.port)
        layout.addWidget(self.username)
        layout.addWidget(self.connect_button)
        layout.addWidget(self.status)

        widget = QWidget()
        widget.setLayout(layout)
        self.setCentralWidget(widget)

    def connect_to_server(self, checked):
        self.status.setText("Connecting...")
        self.status.show()
        try:
            client = Client(server_ip=self.ip.text(), server_port=int(self.port.text()),
                            sign_up_username=self.username.text())
            self.status.hide()
        except:
            self.status.setText("Connection error!")
            return

        self.w = MainWindow(self, client)
        self.hide()
        self.w.show()


def main():
    # logging.basicConfig(level=logging.DEBUG, format='%(asctime)s %(funcName)20s() %(levelname)s %(message)s')
    app = QApplication(sys.argv)
    w = ServerConnectionWindow()
    w.show()
    app.exec()


if __name__ == '__main__':
    main()
