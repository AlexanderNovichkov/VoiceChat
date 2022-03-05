# Voice Chat

## Run server

```
cd server
docker build --tag voice_chat_server .
docker run -p 8081:8081 voice_chat_server        
```

## Run client

```
cd client
python3 -m venv venv
source venv/bin/activate
python3 -m pip install -r requirements.txt
python3 main.py
```