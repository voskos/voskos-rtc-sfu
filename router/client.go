package router

import (
	"io"
	"log"
	"fmt"
	"encoding/json"
	"github.com/pion/webrtc/v3"
	"github.com/gorilla/websocket"
	"github.com/voskos/voskos-rtc-sfu/constant"
)

type Client struct {
	Room *Room
	UserID string
	Conn *websocket.Conn
	PC  *webrtc.PeerConnection
	Audio *webrtc.Track
	Video *webrtc.Track
	Sensor chan constant.RequestBody


}

type SDPBody struct {
    Action string `json:"action"`
    UserID string `json:"user_id"`
    SDP webrtc.SessionDescription `json:"sdp"`
}

func AddClientToRoom(room *Room, user_id string, conn *websocket.Conn, pc *webrtc.PeerConnection) *Client{
	log.Println("[CLIENT] - Added a client to the room")
	client := &Client{Room:  room, UserID : user_id, Conn : conn, PC : pc}
	log.Println("[CLIENT] - Registering the client to the room")
	client.Room.Register <- client
	return client
}

func (self *Client) RenegotiateDueToNewClientJoin(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RENEGOTIATE    )*************************************")
    newJoineeID := reqBody.UserID
    //Loop over other clients in the room and consume tracks
    for client, status := range self.Room.Clients {
        if status {
            //skip my tracks
            if client.UserID == newJoineeID{

                _, err := self.PC.AddTransceiverFromTrack(client.Audio)
                if err != nil {
                    log.Println("Error in consuming audio track of ", client.UserID)
                }

                _, err = self.PC.AddTransceiverFromTrack(client.Video)
                if err != nil {
                    log.Println("Error in consuming video track of ", client.UserID)
                }

                //inititae renegotiation
                // Create offer
                offer, err := self.PC.CreateOffer(nil)
                if err != nil {
                    log.Fatalln(err)
                }

                // Sets the LocalDescription, and starts our UDP listeners
                err = self.PC.SetLocalDescription(offer)
                if err != nil {
                    log.Fatalln(err)
                }

                //Send SDP Answer
                respBody := constant.SDPResponse{}
                respBody.Action = "SERVER_OFFER"
                respBody.SDP = offer
                off, _ := json.Marshal(respBody)
                log.Println("[SENSOR] - SDP Offer Sent")
                self.Conn.WriteMessage(websocket.TextMessage, off)

            }
        }
    }

}

func (my *Client) RenegotiateDueToSelfJoin(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RENEGOTIATE    )*************************************")
    my_id := reqBody.UserID
    //Loop over other clients in the room and consume tracks
    for his, status := range my.Room.Clients {
        if status {
            //skip my tracks
            if his.UserID != my_id{

                _, err := my.PC.AddTransceiverFromTrack(his.Audio)
                if err != nil {
                    log.Println("Error in consuming audio track of ", his.UserID)
                }

            }
        }
    }

    //inititae renegotiation
    // Create offer
    offer, err := my.PC.CreateOffer(nil)
    if err != nil {
        log.Fatalln(err)
    }

    // Sets the LocalDescription, and starts our UDP listeners
    err = my.PC.SetLocalDescription(offer)
    if err != nil {
        log.Fatalln(err)
    }

    //Send SDP Answer
    respBody := constant.SDPResponse{}
    respBody.Action = "SERVER_OFFER"
    respBody.SDP = offer
    off, _ := json.Marshal(respBody)
    log.Println("[SENSOR] - SDP Offer Sent")
    my.Conn.WriteMessage(websocket.TextMessage, off)

}

func (c *Client) Activate() {
	log.Println("[CLIENT] - Client Activated")
	go func() {
		for {
			select {
			case reqBody := <-c.Sensor:
			    action_type := reqBody.Action
			    log.Println("[CLIENT] - Message recieved with action : ", action_type)

			    switch action_type {

				    case "RENEGOTIATE_EXIST_CLIENT":
				        c.RenegotiateDueToNewClientJoin(reqBody)

			     	case "RENEGOTIATE_SELF_CLIENT":
				        c.RenegotiateDueToSelfJoin(reqBody)

				    // case "RENEGOTIATE_NEW_CLIENT":
				    //     sensor.RenegotiateOfNewClient(conn, reqBody)
			    }
			}
		}
	}()
}

func (c *Client) SetAudioTrack(t *webrtc.Track){
	c.Audio = t
	log.Printf("[CLIENT] - Audio track for USER = %s saved with TRACK_ID = %s", c.UserID, c.Audio.ID())
}

func (c *Client) SetVideoTrack(t *webrtc.Track){
	c.Video = t
	log.Printf("[CLIENT] - Video track for USER = %s saved with TRACK_ID = %s", c.UserID, c.Video.ID())
}

func (c *Client) ConsumeAudioTracks(msgType int, pc *webrtc.PeerConnection){
	//Loop over other clients in the room and consume tracks
	for client, status := range c.Room.Clients {
		if status {
			//skip my tracks
			if client.UserID != c.UserID{

				_, err := pc.AddTransceiverFromTrack(client.Audio)
				if err != nil {
					log.Println("Error in consuming audio track of ", client.UserID)
				}

				//inititae renegotiation
				// Create offer
	            offer, err := pc.CreateOffer(nil)
	            if err != nil {
	                log.Fatalln(err)
	            }

	            // Sets the LocalDescription, and starts our UDP listeners
	            err = pc.SetLocalDescription(offer)
	            if err != nil {
	                log.Fatalln(err)
	            }

	            //Send SDP Answer
	            respBody := SDPBody{}
	            respBody.Action = "SERVER_OFFER"
	            respBody.UserID = c.UserID
	            respBody.SDP = offer
	            off, _ := json.Marshal(respBody)
	            log.Println("SDP Offer Sent")
	            c.Conn.WriteMessage(msgType, off)

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