package types

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
	"github/gitalek/go_sandbox_apps/trace/pkg/trace"
	"log"
	"net/http"
)

type Room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	forward chan *message

	// The join and leave channels exist simply to allow us to safely
	// and remove clients from the clients map.
	// join is a channel for clients wishing to join the room.
	join chan *client
	// leave is a channel for clients wishing to leave the room.
	leave chan *client
	// clients holds all current clients in this room.
	clients map[*client]bool
	// tracer will receive trace information of activity in the room
	Tracer trace.Tracer
}

func NewRoom() *Room {
	return &Room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		Tracer: trace.Off(),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
			r.Tracer.Trace("New client joined")
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
			r.Tracer.Trace("Client left")
		case msg := <-r.forward:
			// forward messages to all clients
			for client := range r.clients {
				client.send <- msg
				r.Tracer.Trace("-- sent to client")
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *Room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Upgrade the HTTP server connection to the WebSocket protocol.
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatalf("ServeHTTP: %v", err)
		return
	}
	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatalf("Failed to get auth cookie: %#v\n", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan *message, messageBufferSize),
		room:   r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
