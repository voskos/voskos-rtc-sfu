package action

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/voskos/voskos-rtc-sfu/constant"
	"github.com/voskos/voskos-rtc-sfu/router"
    "github.com/voskos/voskos-rtc-sfu/util"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	// PLI (Pictire Loss Indication)
	rtcpPLIInterval = time.Second * 3
)



//Define actions below
func Init(rtr *router.Router, conn *websocket.Conn, reqBody constant.RequestBody) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	log.Info("*** INIT ***")

	var myRoom *router.Room
	userID := reqBody.UserID
	roomID := reqBody.RoomID
	offer := reqBody.SDP
    body := reqBody.Body
    streamId := body["stream_id"]
    deviceType := body["device_type"]

	log.Info("Init request from user : ", userID, " for room : ", roomID)

	roomExists := false
	for rm, status := range rtr.Rooms {
		if status {
			if rm.RoomID == roomID {
				myRoom = rm
				roomExists = true
				log.Info("Room ID exists in Router context : ", myRoom.RoomID)
				break

			}
		}
	}

	if !roomExists {
		log.Println("Room ID doesn't exists in Router context. Creating a new room in context by : ", userID, " for room :  ", roomID)
		myRoom = router.NewRoom(rtr, roomID)
		go myRoom.Run()
	}

	log.Info("User ", userID, " waiting for the room to be unlocked")
	// for myRoom.IsRoomLocked() {

	// }

	//fmt.Println("LOCKING THE ROOM BY ----", userID)
	//myRoom.LockRoom()
	myRoom.Mu.Lock()
	log.Info("Room lock acquired by ", userID)

	// myRoom.Lock.Lock()
	// Lock locks m. If the lock is already in use, the calling goroutine blocks until the mutex is available.
	// defer myRoom.Lock.Unlock()

	//create a peerconnection object
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		log.Println(err)
        util.SendErrMessage("Failed to create peer connection", conn)

	}

	me := router.AddClientToRoom(myRoom, userID, conn, peerConnection)
	me.Activate()

	peerConnection.OnSignalingStateChange(func(sigState webrtc.SignalingState) {
		log.Info("Signal State ---> ", sigState, " for ", me.UserID)
	})

    me.SaveStreamIdToDeviceTypeInfo(streamId, deviceType)

	// peerConnection.OnNegotiationNeeded(func(){
	//     offer, err := me.PC.CreateOffer(nil)
	//     if err != nil {
	//         log.Println(err)
	//     }

	//     // Sets the LocalDescription, and starts our UDP listeners
	//     err = me.PC.SetLocalDescription(offer)
	//     if err != nil {
	//         log.Println(err)
	//     }

	//     //Send SDP Answer
	//     respBody := constant.SDPResponse{}
	//     respBody.Action = "SERVER_OFFER"
	//     respBody.SDP = offer
	//     off, _ := json.Marshal(respBody)
	//     log.Println("[SENSOR] - SDP Offer Sent to ", me.UserID)
	//     me.Conn.WriteMessage(websocket.TextMessage, off)
	// })

	// peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate){
	//     log.Println("[ACTION - INIT] - NEW ICE CANDIDATE DISCOVERED ---> ", candidate)
	//     //Send SDP Answer
	//     reqBody := constant.ICEResponse{}
	//     reqBody.Action = "NEW_ICE_CANDIDATE_SERVER"
	//     reqBody.ICE_Candidate = candidate
	//     cand, _ := json.Marshal(reqBody)
	//     log.Println("[ACTION - INIT] - ICE Candidate Sent")
	//     conn.WriteMessage(websocket.TextMessage, cand)
	// })

	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
        log.Println("********************************************", remoteTrack.StreamID()  )
        streamId := remoteTrack.StreamID()
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}})
				if errSend != nil {
					log.Error(errSend)
				}
			}
		}()
        if me.StreamIdDeviceTypeMap[streamId] == "webcam"{
            if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
                me.SetAudioTrack(remoteTrack)
            } else {
                log.Println(remoteTrack.StreamID(), "*********************************************")
                me.SetVideoTrack(remoteTrack)
            }
        }else if me.StreamIdDeviceTypeMap[streamId] == "display"{
            if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
                //me.SetAudioTrack(remoteTrack)
            } else {
                log.Println(remoteTrack.StreamID(), "*********************************************")
                me.SetDisplayVideoTrack(remoteTrack)
            }
        }
		
	})

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Println(err)
        util.SendErrMessage("Failed to save remote description", conn)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println(err)
        util.SendErrMessage("Failed to create answer", conn)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println(err)
        util.SendErrMessage("Failed to set local description", conn)
	}

	//Send SDP Answer
	respBody := constant.SDPResponse{}
	respBody.Action = "SERVER_ANSWER"
	respBody.SDP = answer
	ans, _ := json.Marshal(respBody)
	log.Info("Init SDP Answer Sent to ", me.UserID)
	conn.WriteMessage(websocket.TextMessage, ans)

	//Loop over other clients in the room and consume tracks
	log.Info("Init - Room lenght : ", len(me.Room.Clients))
	log.Info("Init - user %s waiting for its video track to get saved", me.UserID)

	//myRoom.UnlockRoom()

}

func RespondToClientAnswer(rtr *router.Router, reqBody constant.RequestBody) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	log.Info("*** Respond to Client Answer ***", reqBody.UserID)

	log.Info("SDP Answer recieved from %s", reqBody.UserID)
	var selfRoom *router.Room
	userID := reqBody.UserID
	roomID := reqBody.RoomID
	answer := reqBody.SDP

	for rm, status := range rtr.Rooms {
		if status {
			if rm.RoomID == roomID {
				selfRoom = rm
				break

			}
		}
	}

	for client, status := range selfRoom.Clients {
		if status {
			if client.UserID == userID {
				// Sets the RemoteDescription
				err := client.PC.SetRemoteDescription(answer)
				log.Info("SDP Answer saved for %s\n", userID)
				if err != nil {
					log.Println(err)
                    util.SendErrMessage("Failed to set remote answer", client.Conn)
				}
				client.PCLock.Unlock()
				log.Info("%s unlocked its PC\n", userID)
				break

			}
		}
	}
}

func StopScreenShare(rtr *router.Router, reqBody constant.RequestBody) {

    log.SetFormatter(&log.TextFormatter{
        FullTimestamp: true,
    })
    log.SetReportCaller(true)

    log.Info("*** Stop screen share ***", reqBody.UserID)

    log.Info("SDP Answer recieved from %s", reqBody.UserID)
    var selfRoom *router.Room
    userID := reqBody.UserID
    roomID := reqBody.RoomID
    ans := reqBody.SDP

    for rm, status := range rtr.Rooms {
        if status {
            if rm.RoomID == roomID {
                selfRoom = rm
                break

            }
        }
    }

    for client, status := range selfRoom.Clients {
        if status {
            if client.UserID == userID {
                // Sets the RemoteDescription)
                client.PCLock.Lock()
                err := client.PC.SetRemoteDescription(ans)
                log.Info("SDP Answer saved for %s\n", userID)
                if err != nil {
                    log.Println(err)
                    util.SendErrMessage("Failed to set remote answer", client.Conn)
                }

                // Create answer
                answer, err := client.PC.CreateAnswer(nil)
                if err != nil {
                    log.Println(err)
                    util.SendErrMessage("Failed to create answer", client.Conn)
                }

                // Sets the LocalDescription, and starts our UDP listeners
                err = client.PC.SetLocalDescription(answer)
                if err != nil {
                    log.Println(err)
                    util.SendErrMessage("Failed to set local description", client.Conn)
                }

                //Send SDP Answer
                respBody := constant.SDPResponse{}
                respBody.Action = "SERVER_ANSWER"
                respBody.SDP = answer
                ans, _ := json.Marshal(respBody)
                log.Info("Init SDP Answer Sent to ", client.UserID)
                client.Conn.WriteMessage(websocket.TextMessage, ans)

                client.PCLock.Unlock()
                log.Info("%s unlocked its PC\n", userID)

            }else{
                //ask others to stop consuming screen share media of this client
                reqBody1 := constant.RequestBody{}
                reqBody1.Action = "SIGNAL_TO_STOP_CONSUME_DISPLAY_VIDEO"
                reqBody1.UserID = userID
                client.SensorDisplayVideo <- reqBody1
            }
        }
    }
}

func RenegotiateScreenShare(rtr *router.Router, reqBody constant.RequestBody) {

    log.SetFormatter(&log.TextFormatter{
        FullTimestamp: true,
    })
    log.SetReportCaller(true)

    log.Info("*** Respond to Client Answer ***", reqBody.UserID)

    log.Info("SDP Answer recieved from %s", reqBody.UserID)
    var selfRoom *router.Room
    userID := reqBody.UserID
    roomID := reqBody.RoomID
    answer := reqBody.SDP
    body := reqBody.Body
    streamId := body["stream_id"]
    deviceType := body["device_type"]

    for rm, status := range rtr.Rooms {
        if status {
            if rm.RoomID == roomID {
                selfRoom = rm
                break

            }
        }
    }

    selfRoom.Mu.Lock()

    for client, status := range selfRoom.Clients {
        if status {
            if client.UserID == userID {
                // Sets the RemoteDescription
                client.PCLock.Lock()
                client.SaveStreamIdToDeviceTypeInfo(streamId, deviceType)
                
                err := client.PC.SetRemoteDescription(answer)
                log.Info("SDP Answer saved for %s\n", userID)
                if err != nil {
                    log.Println(err)
                    util.SendErrMessage("Failed to set remote answer", client.Conn)
                }

                // Create answer
                answer, err := client.PC.CreateAnswer(nil)
                if err != nil {
                    log.Println(err)
                    util.SendErrMessage("Failed to create answer", client.Conn)
                }

                // Sets the LocalDescription, and starts our UDP listeners
                err = client.PC.SetLocalDescription(answer)
                if err != nil {
                    log.Println(err)
                    util.SendErrMessage("Failed to set local description", client.Conn)
                }

                //Send SDP Answer
                respBody := constant.SDPResponse{}
                respBody.Action = "SERVER_ANSWER"
                respBody.SDP = answer
                ans, _ := json.Marshal(respBody)
                log.Info("Init SDP Answer Sent to ", client.UserID)
                client.Conn.WriteMessage(websocket.TextMessage, ans)


                client.PCLock.Unlock()
                log.Info("%s unlocked its PC\n", userID)
                break

            }
        }
    }
}

func AddIceCandidate(rtr *router.Router, reqBody constant.RequestBody) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	
	log.Info("*** Adding ICE Candidate ***")

	var selfRoom *router.Room
	userID := reqBody.UserID
	roomID := reqBody.RoomID
	ice_candidate := reqBody.ICE_Candidate.ToJSON()
	log.Info("[ACTION] - New ICECandidate %v recieved from %s", ice_candidate, userID)
	//ToJSON returns an ICECandidateInit which is used in AddIceCandidate

	for rm, status := range rtr.Rooms {
		if status {
			if rm.RoomID == roomID {
				selfRoom = rm
				break

			}
		}
	}

	for client, status := range selfRoom.Clients {
		if status {
			if client.UserID == userID {

				// Sets the RemoteDescription
				err := client.PC.AddICECandidate(ice_candidate)
				if err != nil {
					log.Println(err)
                    util.SendErrMessage("Failed to add ICE candidate", client.Conn)
				}

				break

			}
		}
	}
}
