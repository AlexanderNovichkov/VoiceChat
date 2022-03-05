FROM golang:1.17-alpine

WORKDIR /server

COPY . .

RUN go mod download
RUN go build -o /app

EXPOSE 8081

CMD [ "/app" ]