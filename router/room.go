package router

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

type Room struct {
	Mu         sync.Mutex
	Lock       bool
	Router     *Router
	RoomID     string
	Clients    map[*Client]bool // Registered clients.
	Register   chan *Client     // Register requests from the clients.
	Unregister chan *Client     // Unregister requests from clients.
}

func NewRoom(router *Router, roomID string) *Room {
	room := &Room{
		Mu:         sync.Mutex{},
		Router:     router,
		RoomID:     roomID,
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
	room.Router.Register <- room
	return room
}

func (r *Room) Run() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	
	for {
		select {
		case client := <-r.Register:
			log.Info("User ID ", client.UserID, " registered in room ID ", r.RoomID)
			r.Clients[client] = true
		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
			}
		}
	}
}

func (r *Room) LockRoom() {
	r.Lock = true
}

func (r *Room) UnlockRoom() {
	r.Lock = false
}

func (r *Room) IsRoomLocked() bool {
	return r.Lock
}
