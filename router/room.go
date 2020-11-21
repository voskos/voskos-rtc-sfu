package router

import (
	"log"
	"sync"
)
type Room struct {
	Lock sync.Mutex
	Router *Router
	RoomID string
	Clients map[*Client]bool // Registered clients.
	Register chan *Client // Register requests from the clients.
	Unregister chan *Client // Unregister requests from clients.
}

func NewRoom(router *Router, roomID string) *Room {
	room := &Room{
		Router : router,
		RoomID : roomID,
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
	room.Router.Register <- room
	return room
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			log.Printf("[ROOM] - User %s registered in room", client.UserID)
			r.Clients[client] = true
		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
			}
		}
	}
}
