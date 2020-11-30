package router

import (
	log "github.com/sirupsen/logrus"
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
		Rooms:      make(map[*Room]bool),
	}
}

func (r *Router) Run() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	log.Info("Router is up and running")
	for {
		select {
		case room := <-r.Register:
			log.Info("Room ID %s registered in router's room context", room.RoomID)
			r.Rooms[room] = true
		case room := <-r.Unregister:
			if _, ok := r.Rooms[room]; ok {
				delete(r.Rooms, room)
			}
		}
	}
}
