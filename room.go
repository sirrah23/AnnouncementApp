package main

type room struct {
	name          string
	announcements []string
	forward       chan []byte
	join          chan *client
	leave         chan *client
	clients       map[*client]bool
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

type roomManager struct {
	rooms map[string]*room
}

func (rm *roomManager) getRoom(name string) (bool, *room) {
	room, ok := rm.rooms[name]
	if !ok {
		return false, nil
	} else {
		return true, room
	}
}

func (rm *roomManager) addRoom(name string, room *room) {
	rm.rooms[name] = room
}

func newRoomManager() *roomManager {
	return &roomManager{
		rooms: make(map[string]*room),
	}
}
