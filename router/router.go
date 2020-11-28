package router

import (
	"log"
)
type Router struct {
	// Registered rooms.
	Rooms map[*Room]bool

	// Register requests for rooms
	Register chan *Room

	// Unregister requests for rooms
	Unregister chan *Room
}

func NewRouter() *Router {
	return &Router{
		Register:   make(chan *Room),
		Unregister: make(chan *Room),
		Rooms:    make(map[*Room]bool),
	}
}

func (r *Router) Run() {
	log.Println("[ROUTER] - Router is up and running")
	for {
		select {
		case room := <-r.Register:
			log.Printf("[ROUTER] - Room %s registered in room", room.RoomID)
			r.Rooms[room] = true
		case room := <-r.Unregister:
			if _, ok := r.Rooms[room]; ok {
				delete(r.Rooms, room)
			}
		}
	}
}
