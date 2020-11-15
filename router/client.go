package router

import (
	"log"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Client struct {
	Room *Room
	UserID string
	// The websocket connection.
	Conn *websocket.Conn
	PC  *webrtc.PeerConnection

}


func AddClientToRoom(room *Room, user_id string, conn *websocket.Conn, pc *webrtc.PeerConnection) {
	log.Println("Added a client to the room")
	client := &Client{Room:  room, UserID : user_id, Conn : conn, PC : pc}
	log.Println("Registering the client to the room")
	client.Room.Register <- client
}