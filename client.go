package main

import (
	"github.com/gorilla/websocket"
)

type client struct {
	socket *websocket.Conn
	send   chan []byte
	room   *room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, announcement, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- announcement
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for announcement := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, announcement)
		if err != nil {
			return
		}
	}
}
