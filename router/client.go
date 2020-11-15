package router

import (
	"io"
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
	Audio *webrtc.Track
	Video *webrtc.Track

}


func AddClientToRoom(room *Room, user_id string, conn *websocket.Conn, pc *webrtc.PeerConnection) *Client{
	log.Println("Added a client to the room")
	client := &Client{Room:  room, UserID : user_id, Conn : conn, PC : pc}
	log.Println("Registering the client to the room")
	client.Room.Register <- client
	return client
}

func (c *Client) SetAudioTrack(t *webrtc.Track){
	c.Audio = t
	log.Printf("Audio track for USER = %s set with TRACK_ID = %s", c.UserID, c.Audio.ID())
}

func (c *Client) SetVideoTrack(t *webrtc.Track){
	c.Video = t
	log.Printf("Video track for USER = %s set with TRACK_ID = %s", c.UserID, c.Video.ID())
}

func (c *Client) ConsumeAudioTracks(pc *webrtc.PeerConnection){
	//Loop over other clients in the room and consume tracks
	for client, status := range c.Room.Clients {
		if status {
			//skip my tracks
			if client.UserID != c.UserID{

				// Create a local track, all our SFU clients will be fed via this track
				localTrack, newTrackErr := pc.NewTrack(client.Audio.PayloadType(), client.Audio.SSRC(), "audio", client.UserID)
				if newTrackErr != nil {
					log.Println("Error in consuming audio track of ", client.UserID)
				}

				rtpBuf := make([]byte, 1400)
				for {
					i, readErr := client.Audio.Read(rtpBuf)
					if readErr != nil {
						log.Fatalln(readErr)
					}

					// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
					if _, err := localTrack.Write(rtpBuf[:i]); err != nil && err != io.ErrClosedPipe {
						log.Fatalln(err)
					}
				}
			}
		}
	}
}

func (c *Client) ConsumeVideoTracks(pc *webrtc.PeerConnection){
	//Loop over other clients in the room and consume tracks
	for client, status := range c.Room.Clients {
		if status {
			//skip my tracks
			if client.UserID != c.UserID{

				// Create a local track, all our SFU clients will be fed via this track
				localTrack, newTrackErr := pc.NewTrack(client.Video.PayloadType(), client.Video.SSRC(), "video", client.UserID)
				if newTrackErr != nil {
					log.Println("Error in consuming video track of ", client.UserID)
				}

				rtpBuf := make([]byte, 1400)
				for {
					i, readErr := client.Video.Read(rtpBuf)
					if readErr != nil {
						log.Fatalln(readErr)
					}

					// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
					if _, err := localTrack.Write(rtpBuf[:i]); err != nil && err != io.ErrClosedPipe {
						log.Fatalln(err)
					}
				}
			}
		}
	}
}