package types

import (
	"github.com/gorilla/websocket"
)

type client struct {
	// socket is the web socket for this client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan []byte
	// room is the room this client is chatting in.
	room *Room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	// The write method of our client type will pick up the message
	// and send it down the socket to the browser.
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
