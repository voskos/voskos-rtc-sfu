package router

import (
	"log"
	"fmt"
	// "time"
	"reflect"
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
	Audio *webrtc.TrackRemote
	Video *webrtc.TrackRemote
	Sensor chan constant.RequestBody
	AudioLock bool
	VideoLock bool


}

func AddClientToRoom(room *Room, user_id string, conn *websocket.Conn, pc *webrtc.PeerConnection) *Client{
	log.Println("[CLIENT] - Added a client to the room")
	client := &Client{
				Room:  room, 
				UserID : user_id, 
				Conn : conn, 
				PC : pc, 
				Audio : nil,
				Video : nil,
				Sensor : make(chan constant.RequestBody),
				AudioLock : true,
				VideoLock : true,
			}
	log.Println("[CLIENT] - Registering the client to the room")
	client.Room.Register <- client
	return client
}

func writeRTPToTrack(outputTrack *webrtc.TrackLocalStaticRTP, track *webrtc.TrackRemote){
	for {
			// Read RTP packets being sent to Pion
			rtp, readErr := track.ReadRTP()
			if readErr != nil {
				panic(readErr)
			}

			if writeErr := outputTrack.WriteRTP(rtp); writeErr != nil {
				panic(writeErr)
			}
		}
}

func (self *Client) RenegotiateDueToNewClientJoin(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RENEGOTIATE - OTHER   )*************************************")
    newJoineeID := reqBody.UserID
    //Loop over other clients in the room and consume tracks
    for client, status := range self.Room.Clients {
        if status {
            //skip my tracks
            if client.UserID == newJoineeID{

            	if client.Video != nil{
            		outputTrackVideo, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", client.UserID)
					if err != nil {
						panic(err)
					}

					// Add this newly created track to the PeerConnection
					if _, err = self.PC.AddTrack(outputTrackVideo); err != nil {
						log.Println("[CLIENT] - Error in adding video output track ", client.UserID)
						panic(err)
					}
					go writeRTPToTrack(outputTrackVideo, client.Video)
            	}

            	if client.Audio != nil{
            		outputTrackAudio, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", client.UserID)
					if err != nil {
						panic(err)
					}

					// Add this newly created track to the PeerConnection
					if _, err = self.PC.AddTrack(outputTrackAudio); err != nil {
						log.Println("[CLIENT] - Error in adding output audio track ", client.UserID)
						panic(err)
					}

					go writeRTPToTrack(outputTrackAudio, client.Audio)
            	}
            	

                break;

            }


        }
    }

    // time.Sleep(3 * time.Second) 
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
    log.Println("[SENSOR] - SDP Offer Sent to ", self.UserID)
    self.Conn.WriteMessage(websocket.TextMessage, off)

}

func (my *Client) RenegotiateDueToSelfJoin(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RENEGOTIATE - SELF  )*************************************")
    my_id := reqBody.UserID
    //Loop over other clients in the room and consume tracks
    for his, status := range my.Room.Clients {
        if status {
            //skip my tracks
            if his.UserID != my_id{

            	if his.Video != nil{
            		// Create Track that we send video back to browser on
	            	log.Println("[TYPE OF TRACK] - ", reflect.TypeOf(his.Audio))
					outputTrackVideo, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", his.UserID)
					if err != nil {
						panic(err)
					}

					// Add this newly created track to the PeerConnection
					if _, err = my.PC.AddTrack(outputTrackVideo); err != nil {
						log.Println("[CLIENT] - Error in adding output track", his.UserID)
						panic(err)
					}
					go writeRTPToTrack(outputTrackVideo, his.Video)
            	}

            	if his.Audio != nil{
            		outputTrackAudio, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", his.UserID)
					if err != nil {
						panic(err)
					}

					// Add this newly created track to the PeerConnection
					if _, err = my.PC.AddTrack(outputTrackAudio); err != nil {
						log.Println("[CLIENT] - Error in adding output audio track ", his.UserID)
						panic(err)
					}

					go writeRTPToTrack(outputTrackAudio, his.Audio)
            	}
            

            }
        }
    } 
    //inititae renegotiation
    // Create offer
    //time.Sleep(3 * time.Second)
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
    log.Println("[SENSOR] - SDP Offer Sent to", my.UserID)
    my.Conn.WriteMessage(websocket.TextMessage, off)

}

func (c *Client) Activate() {
	log.Println("[CLIENT] - Client Activated")
	go func() {
		for {
			select {
			case reqBody := <-c.Sensor:
			    action_type := reqBody.Action
			    log.Println("[CLIENT] - Message recieved with action : ", action_type, " for ", reqBody.UserID)

			    switch action_type {

				    case "RENEGOTIATE_EXIST_CLIENT":
				        go c.RenegotiateDueToNewClientJoin(reqBody)

			     	case "RENEGOTIATE_SELF_CLIENT":
				        c.RenegotiateDueToSelfJoin(reqBody)
			    }
			}
		}
	}()
}

func (c *Client) SetAudioTrack(t *webrtc.TrackRemote){
	c.Audio = t
	c.AudioLock = false
	log.Printf("[CLIENT] - Audio track for USER = %s saved with TRACK_ID = %s", c.UserID, c.Audio.ID())
}

func (c *Client) SetVideoTrack(t *webrtc.TrackRemote){
	c.Video = t
	c.VideoLock = false
	fmt.Println("UNLOCKING ROOM BY ----", c.UserID)
	c.Room.UnlockRoom()
	log.Printf("[CLIENT] - Video track for USER = %s saved with TRACK_ID = %s", c.UserID, c.Video.ID())
}
