package router

import (
	"log"
	"fmt"
	"sync"
	// "time"
	"reflect"
	"encoding/json"
	"github.com/pion/webrtc/v3"
	"github.com/gorilla/websocket"
	"github.com/voskos/voskos-rtc-sfu/constant"
)

type Client struct {
	PCLock sync.Mutex
	Room *Room
	UserID string
	Conn *websocket.Conn
	PC  *webrtc.PeerConnection
	Audio *webrtc.TrackRemote
	Video *webrtc.TrackRemote
	SensorAudio chan constant.RequestBody
	SensorVideo chan constant.RequestBody


}

func AddClientToRoom(room *Room, user_id string, conn *websocket.Conn, pc *webrtc.PeerConnection) *Client{
	client := &Client{
				PCLock : sync.Mutex{},
				Room:  room, 
				UserID : user_id, 
				Conn : conn, 
				PC : pc, 
				Audio : nil,
				Video : nil,
				SensorAudio : make(chan constant.RequestBody),
				SensorVideo: make(chan constant.RequestBody),
			}
	log.Printf("[CLIENT] - Registering %s in %s\n", user_id, room.RoomID)
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

func (self *Client) RenegotiateDueToNewClientJoinAudio(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RENEGOTIATE - OTHER AUDIO)*************************************")
	self.PCLock.Lock()
    log.Printf("%s locked its PC\n", self.UserID)
    newJoineeID := reqBody.UserID
    //Loop over other clients in the room and consume tracks
    for client, status := range self.Room.Clients {
        if status {
            //skip my tracks
            if client.UserID == newJoineeID{

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

func (self *Client) RenegotiateDueToNewClientJoinVideo(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RENEGOTIATE - OTHER VIDEO )*************************************")
	self.PCLock.Lock()
    log.Printf("%s locked its PC\n", self.UserID)
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
            	}else{
            		log.Printf("%s could not consume tracks of %s\n", self.UserID, newJoineeID)
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

func (my *Client) RenegotiateDueToSelfJoinAudio(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RENEGOTIATE - SELF AUDIO )*************************************")
	my.PCLock.Lock()
    log.Printf("%s locked its PC\n", my.UserID)
    my_id := reqBody.UserID
    //Loop over other clients in the room and consume tracks
    for his, status := range my.Room.Clients {
        if status {
            //skip my tracks
            if his.UserID != my_id{

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
    //time.Sleep(3 * time.Second
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

func (my *Client) RenegotiateDueToSelfJoinVideo(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RENEGOTIATE - SELF VIDEO )*************************************")
	my.PCLock.Lock()
    log.Printf("%s locked its PC\n", my.UserID)
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
            	}else{
            		log.Printf("%s could not consume tracks of %s\n", my.UserID, his.UserID)
            	}           

            }
        }
    } 
    //inititae renegotiation
    // Create offer
    //time.Sleep(3 * time.Second
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
	log.Printf("[CLIENT] - Client %s Activated\n", c.UserID)
	go func() {
		for {
			select {
			case reqBody := <-c.SensorAudio:
			    action_type := reqBody.Action
			    log.Println("[CLIENT SENSOR AUDIO] - Message recieved with action : ", action_type, " for ", reqBody.UserID)

			    switch action_type {

				    case "RENEGOTIATE_EXIST_CLIENT":
				        go c.RenegotiateDueToNewClientJoinAudio(reqBody)

			     	case "RENEGOTIATE_SELF_CLIENT":
				        go c.RenegotiateDueToSelfJoinAudio(reqBody)
			    }

			case reqBody := <-c.SensorVideo:
			    action_type := reqBody.Action
			    log.Println("[CLIENT SENSOR VIDEO] - Message recieved with action : ", action_type, " for ", reqBody.UserID)

			    switch action_type {

				    case "RENEGOTIATE_EXIST_CLIENT":
				        go c.RenegotiateDueToNewClientJoinVideo(reqBody)

			     	case "RENEGOTIATE_SELF_CLIENT":
				        go c.RenegotiateDueToSelfJoinVideo(reqBody)
			    }
			}
		}
	}()
}

func (c *Client) SetAudioTrack(t *webrtc.TrackRemote){
	c.Audio = t
	log.Printf("[CLIENT] - Audio track for USER = %s saved with TRACK_ID = %s", c.UserID, c.Audio.ID())

	if len(c.Room.Clients) > 1{
    	for he, status := range c.Room.Clients {
			if status {
				if he.UserID != c.UserID{

		            //Send SDP Answer
		            reqBody := constant.RequestBody{}
		            reqBody.Action = "RENEGOTIATE_EXIST_CLIENT"
		            reqBody.UserID = c.UserID
		            he.SensorAudio <- reqBody

				}else{
					//Send SDP Answer
		            reqBody := constant.RequestBody{}
		            reqBody.Action = "RENEGOTIATE_SELF_CLIENT"
		            reqBody.UserID = c.UserID
		            c.SensorAudio <- reqBody
				}
			}
            // time.Sleep(3 * time.Second) 
		}
    }
}

func (c *Client) SetVideoTrack(t *webrtc.TrackRemote){
	c.Video = t
	log.Printf("[CLIENT] - Video track for USER = %s saved with TRACK_ID = %s\n", c.UserID, c.Video.ID())
	//c.Room.UnlockRoom()
	
	log.Printf("[CLIENT] - Room unlocked by %s\n", c.UserID)
	//If ypur video is saved then send it to others and consume other's video too
	if len(c.Room.Clients) > 1{
		i := 0
    	for he, status := range c.Room.Clients {
    		i++
			if status {
				if he.UserID != c.UserID{

		            //Send SDP Answer
		            reqBody := constant.RequestBody{}
		            reqBody.Action = "RENEGOTIATE_EXIST_CLIENT"
		            reqBody.UserID = c.UserID
		            he.SensorVideo <- reqBody

				}else{
					//Send SDP Answer
		            reqBody := constant.RequestBody{}
		            reqBody.Action = "RENEGOTIATE_SELF_CLIENT"
		            reqBody.UserID = c.UserID
		            c.SensorVideo <- reqBody
				}
			}
			//before unlocking the room, make sure u have interacted with every other client present
			if i == len(c.Room.Clients){
				c.Room.Mu.Unlock()
			}
            // time.Sleep(3 * time.Second) 
		}
    }else{
    	c.Room.Mu.Unlock()
    }
    
}
