package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var (
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  socketBufferSize,
		WriteBufferSize: socketBufferSize,
	}
	rm = newRoomManager()
)

func joinRoom(w http.ResponseWriter, r *http.Request) {
	log.Println("[S] Join room started")
	name := r.URL.Path[len("/joinroom/"):]

	log.Println("[S] Retrieving room")
	success, room := rm.getRoom(name)

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

func createRoom(w http.ResponseWriter, r *http.Request) {
	roomname := r.FormValue("name")
	roomlink := rm.roomNameGenerate(slug.Make(roomname))
	log.Println(fmt.Sprintf("[S] Creating room %s", roomname))
	newroom := newRoom(roomname)
	go newroom.run()
	rm.addRoom(roomlink, newroom)
	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomlink), http.StatusFound)
	log.Println(fmt.Sprintf("[S] Room %s created", roomname))
}

type roomExistanceWrapper struct {
	th *templateHandler
}

func (rW *roomExistanceWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	room := path[len(path)-1]
	ok, _ := rm.getRoom(room)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 Not Found")
		return
	}
	rW.th.ServeHTTP(w, r)
}
