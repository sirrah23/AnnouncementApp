package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
)

type room struct {
	name          string
	announcements []string
	forward       chan []byte
	join          chan *client
	leave         chan *client
	clients       map[*client]bool
}

func newRoom(name string) *room {
	return &room{
		name:          name,
		announcements: make([]string, 0),
		forward:       make(chan []byte),
		join:          make(chan *client),
		leave:         make(chan *client),
		clients:       make(map[*client]bool),
	}
}

func joinRoom(w http.ResponseWriter, r *http.Request) {
	log.Println("[S] Join room started")
	name := r.URL.Path[len("/joinroom/"):]

	log.Println("[S] Retrieving room")
	success, room := getRoom(name)

	log.Println("[S] Checking success")
	if !success {
		http.Error(w, fmt.Sprintf("Error when trying to join room %s", name), http.StatusNotFound)
		return
	}

	log.Println("[S] Upgrading websocket")
	socket, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	log.Println("[S] Creating client")
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   room,
	}

	log.Println("[S] Adding client to room")
	room.join <- client
	defer func() { room.leave <- client }()
	go client.write()
	for _, a := range room.announcements {
		client.send <- []byte(a)
	}
	client.read()
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
		case announcement := <-r.forward:
			r.announcements = append(r.announcements, string(announcement))
			for client := range r.clients {
				client.send <- announcement
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

var allrooms map[string]*room = make(map[string]*room)

func createRoom(w http.ResponseWriter, r *http.Request) {
	roomName := slug.Make(r.FormValue("name"))
	log.Println(fmt.Sprintf("[S] Creating room %s", roomName))
	newroom := newRoom(roomName)
	go newroom.run()
	allrooms[roomName] = newroom //TODO: Add check for room existence
	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusFound)
	log.Println(fmt.Sprintf("[S] Room %s created", roomName))
}

func getRoom(name string) (bool, *room) {
	room, ok := allrooms[name]
	if !ok {
		return false, nil
	} else {
		return true, room
	}
}
