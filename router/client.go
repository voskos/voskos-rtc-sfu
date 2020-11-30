package router

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/voskos/voskos-rtc-sfu/constant"
	"reflect"
)

type Client struct {
	PCLock      sync.Mutex
	Room        *Room
	UserID      string
	Conn        *websocket.Conn
	PC          *webrtc.PeerConnection
	Audio       *webrtc.TrackRemote
	Video       *webrtc.TrackRemote
	SensorAudio chan constant.RequestBody
	SensorVideo chan constant.RequestBody
}

func AddClientToRoom(room *Room, user_id string, conn *websocket.Conn, pc *webrtc.PeerConnection) *Client {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	
	client := &Client{
		PCLock:      sync.Mutex{},
		Room:        room,
		UserID:      user_id,
		Conn:        conn,
		PC:          pc,
		Audio:       nil,
		Video:       nil,
		SensorAudio: make(chan constant.RequestBody),
		SensorVideo: make(chan constant.RequestBody),
	}
	log.Info("Registering user", user_id, " in room ", room.RoomID)
	client.Room.Register <- client
	return client
}

func writeRTPToTrack(outputTrack *webrtc.TrackLocalStaticRTP, track *webrtc.TrackRemote) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

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

func (self *Client) RenegotiateDueToNewClientJoinAudio(reqBody constant.RequestBody) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	log.Info("*** Renegotiating [other's] audio ***")
	self.PCLock.Lock()
	log.Info("User %s locked its Peer Connection Object", self.UserID)
	newJoineeID := reqBody.UserID
	//Loop over other clients in the room and consume tracks
	for client, status := range self.Room.Clients {
		if status {
			//skip my tracks
			if client.UserID == newJoineeID {

				if client.Audio != nil {
					outputTrackAudio, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", client.UserID)
					if err != nil {
						panic(err)
					}

					// Add this newly created track to the PeerConnection
					if _, err = self.PC.AddTrack(outputTrackAudio); err != nil {
						log.Error("Error in adding output audio track for user ", client.UserID)
						panic(err)
					}

					go writeRTPToTrack(outputTrackAudio, client.Audio)
				}

				break

			}

		}
	}

	// time.Sleep(3 * time.Second)
	//inititae renegotiation
	// Create offer
	offer, err := self.PC.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = self.PC.SetLocalDescription(offer)
	if err != nil {
		log.Fatal(err)
	}

	//Send SDP Answer
	respBody := constant.SDPResponse{}
	respBody.Action = "SERVER_OFFER"
	respBody.SDP = offer
	off, _ := json.Marshal(respBody)
	log.Info("[Sensor] SDP Offer Sent to user ", self.UserID)
	self.Conn.WriteMessage(websocket.TextMessage, off)

}

func (self *Client) RenegotiateDueToNewClientJoinVideo(reqBody constant.RequestBody) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	
	log.Info("*** Renegotiating [other's] video ***")
	self.PCLock.Lock()
	log.Info("User %s locked its Peer Connection", self.UserID)
	newJoineeID := reqBody.UserID
	//Loop over other clients in the room and consume tracks
	for client, status := range self.Room.Clients {
		if status {
			//skip my tracks
			if client.UserID == newJoineeID {

				if client.Video != nil {
					outputTrackVideo, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", client.UserID)
					if err != nil {
						panic(err)
					}

					// Add this newly created track to the PeerConnection
					if _, err = self.PC.AddTrack(outputTrackVideo); err != nil {
						log.Error("Error in adding video output track for user ", client.UserID)
						panic(err)
					}
					go writeRTPToTrack(outputTrackVideo, client.Video)
				} else {
					log.Info("User %s could not consume tracks of user %s\n", self.UserID, newJoineeID)
				}

				break

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

func (my *Client) RenegotiateDueToSelfJoinAudio(reqBody constant.RequestBody) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	
	log.Info("*** Renegotiating [self] audio ***")
	my.PCLock.Lock()
	log.Info("User %s locked its Peer Connection\n", my.UserID)
	my_id := reqBody.UserID
	//Loop over other clients in the room and consume tracks
	for other, status := range my.Room.Clients {
		if status {
			//skip my tracks
			if other.UserID != my_id {

				if other.Audio != nil {
					outputTrackAudio, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", other.UserID)
					if err != nil {
						panic(err)
					}

					// Add this newly created track to the PeerConnection
					if _, err = my.PC.AddTrack(outputTrackAudio); err != nil {
						log.Error("Error in adding output audio track for other user ", other.UserID)
						panic(err)
					}

					go writeRTPToTrack(outputTrackAudio, other.Audio)
				}

			}
		}
	}
	//inititae renegotiation
	// Create offer
	//time.Sleep(3 * time.Second
	offer, err := my.PC.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = my.PC.SetLocalDescription(offer)
	if err != nil {
		log.Fatal(err)
	}

	//Send SDP Answer
	respBody := constant.SDPResponse{}
	respBody.Action = "SERVER_OFFER"
	respBody.SDP = offer
	off, _ := json.Marshal(respBody)
	log.Info("[SENSOR] - SDP Offer Sent to user ", my.UserID)
	my.Conn.WriteMessage(websocket.TextMessage, off)

}

func (my *Client) RenegotiateDueToSelfJoinVideo(reqBody constant.RequestBody) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	log.Info("*** Renegotiating [self] video ***")
	my.PCLock.Lock()
	log.Info("User %s locked its Peer Connection\n", my.UserID)
	my_id := reqBody.UserID
	//Loop over other clients in the room and consume tracks
	for other, status := range my.Room.Clients {
		if status {
			//skip my tracks
			if other.UserID != my_id {

				if other.Video != nil {
					// Create Track that we send video back to browser on
					log.Info("[TYPE OF TRACK] - ", reflect.TypeOf(other.Audio))
					outputTrackVideo, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", other.UserID)
					if err != nil {
						panic(err)
					}

					// Add this newly created track to the PeerConnection
					if _, err = my.PC.AddTrack(outputTrackVideo); err != nil {
						log.Error("Error in adding output track", other.UserID)
						panic(err)
					}
					go writeRTPToTrack(outputTrackVideo, other.Video)
				} else {
					log.Info("User %s could not consume tracks of user %s\n", my.UserID, other.UserID)
				}

			}
		}
	}
	//inititae renegotiation
	// Create offer
	//time.Sleep(3 * time.Second
	offer, err := my.PC.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = my.PC.SetLocalDescription(offer)
	if err != nil {
		log.Fatal(err)
	}

	//Send SDP Answer
	respBody := constant.SDPResponse{}
	respBody.Action = "SERVER_OFFER"
	respBody.SDP = offer
	off, _ := json.Marshal(respBody)
	log.Info("[SENSOR] - SDP Offer Sent to user", my.UserID)
	my.Conn.WriteMessage(websocket.TextMessage, off)

}

func (c *Client) Activate() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	log.Info("Client ", c.UserID," Activated")
	go func() {
		for {
			select {
			case reqBody := <-c.SensorAudio:
				action_type := reqBody.Action
				log.Info("[CLIENT SENSOR AUDIO] - Message recieved with action : ", action_type, " for ", reqBody.UserID)

				switch action_type {

				case "RENEGOTIATE_EXIST_CLIENT":
					go c.RenegotiateDueToNewClientJoinAudio(reqBody)

				case "RENEGOTIATE_SELF_CLIENT":
					go c.RenegotiateDueToSelfJoinAudio(reqBody)
				}

			case reqBody := <-c.SensorVideo:
				action_type := reqBody.Action
				log.Info("[CLIENT SENSOR VIDEO] - Message recieved with action : ", action_type, " for ", reqBody.UserID)

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

func (c *Client) SetAudioTrack(t *webrtc.TrackRemote) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	c.Audio = t
	log.Info("[CLIENT] Audio track for USER : %s saved with TRACK_ID : %s", c.UserID, c.Audio.ID())

	if len(c.Room.Clients) > 1 {
		for he, status := range c.Room.Clients {
			if status {
				if he.UserID != c.UserID {

					//Send SDP Answer
					reqBody := constant.RequestBody{}
					reqBody.Action = "RENEGOTIATE_EXIST_CLIENT"
					reqBody.UserID = c.UserID
					he.SensorAudio <- reqBody

				} else {
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

func (c *Client) SetVideoTrack(t *webrtc.TrackRemote) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	
	c.Video = t
	log.Info("[CLIENT] Video track for USER = %s saved with TRACK_ID = %s\n", c.UserID, c.Video.ID())
	//c.Room.UnlockRoom()

	log.Info("[CLIENT] - Room unlocked by %s\n", c.UserID)
	//If ypur video is saved then send it to others and consume other's video too
	if len(c.Room.Clients) > 1 {
		i := 0
		for other, status := range c.Room.Clients {
			i++
			if status {
				if other.UserID != c.UserID {

					//Send SDP Answer
					reqBody := constant.RequestBody{}
					reqBody.Action = "RENEGOTIATE_EXIST_CLIENT"
					reqBody.UserID = c.UserID
					other.SensorVideo <- reqBody

				} else {
					//Send SDP Answer
					reqBody := constant.RequestBody{}
					reqBody.Action = "RENEGOTIATE_SELF_CLIENT"
					reqBody.UserID = c.UserID
					c.SensorVideo <- reqBody
				}
			}
			//before unlocking the room, make sure u have interacted with every other client present
			if i == len(c.Room.Clients) {
				c.Room.Mu.Unlock()
			}
			// time.Sleep(3 * time.Second)
		}
	} else {
		c.Room.Mu.Unlock()
	}

}
