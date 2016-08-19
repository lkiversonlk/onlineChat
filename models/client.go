package models

import (
	"github.com/gorilla/websocket"
	"time"
)

type Client struct {
	socket   *websocket.Conn
	send     chan *message
	room     *Room
	UserData map[string]interface{}
}

func (c *Client)read() {
	for {
		var msg *message

		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.UserData["name"].(string)

			msg.AvatarURL, _ = c.room.avatar.GetAvatarURL(c)
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *Client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}