package models

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"github.com/lkiversonlk/OnlineChat/trace"
	"github.com/mesos/mesos-go/examples/Godeps/_workspace/src/github.com/stretchr/objx"
)

type Room struct {
	forward chan *message
	join    chan *Client
	leave   chan *Client
	clients map[*Client]bool
	Tracer  trace.Tracer
	avatar  Avatar
}

func NewRoom(avatar Avatar) *Room {
	return &Room{
		forward: make(chan *message),
		join: make(chan *Client),
		leave: make(chan *Client),
		clients: make(map[*Client]bool),
		Tracer: trace.NilTracer(),
		avatar: avatar,
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.Tracer.Trace("New client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.Tracer.Trace("Client left")
		case msg := <-r.forward:
			for client := range r.clients {
				select {
				case client.send <- msg:
					r.Tracer.Trace(" - - sent to client")
				default:
					delete(r.clients, client)
					close(client.send)
					r.Tracer.Trace(" - - failed to send, cleaned up client")
				}
			}
		}
	}
}

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *Room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	authCookie, err := req.Cookie("auth")

	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}

	client := &Client{
		socket: socket,
		send: make(chan *message, messageBufferSize),
		room: r,
		UserData: objx.MustFromBase64(authCookie.Value),
	}

	r.join <- client
	defer func() {
		r.leave <- client
	}()
	go client.write()
	client.read()
}