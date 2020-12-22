package types

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type client struct {
	// socket is the web socket for this client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan *message
	// room is the room this client is chatting in.
	room *Room
	// userData holds information about the user
	userData map[string]interface{}
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("client.read err: %#v", err)
			return
		}
		msg.When = time.Now()
		msg.Name = c.userData["name"].(string)
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	// The write method of our client type will pick up the message
	// and send it down the socket to the browser.
	for msg := range c.send {
		err := c.socket.WriteJSON(msg)
		if err != nil {
			fmt.Printf("client.write err: %#v", err)
			return
		}
	}
}
