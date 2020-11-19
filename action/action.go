package action

import(
	"fmt"
	"log"
	"time"
	"encoding/json"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/gorilla/websocket"
	"github.com/voskos/voskos-rtc-sfu/constant"
	"github.com/voskos/voskos-rtc-sfu/router"
)

const (
    // PLI (Pictire Loss Indication)
    rtcpPLIInterval = time.Second * 3
)

var flag int = 1
var room *router.Room

//Define actions below
func Init(conn *websocket.Conn, reqBody constant.RequestBody){
	fmt.Println("***************************************************(   INIT    )*************************************")
	userID := reqBody.UserID
	roomID := reqBody.RoomID
	offer := reqBody.SDP
	log.Println("[ACTION - INIT] - Init request from ", userID , " for ", roomID)

	if flag == 1{
		flag = 0
		room = router.NewRoom()
		go room.Run()
	}

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
        log.Fatalln(err)
    }

    peerConnection.OnSignalingStateChange(func(sigState webrtc.SignalingState){
        log.Println("[ACTION - INIT] - SIGNAL STATE ---> ", sigState)
    })

    me := router.AddClientToRoom(room, userID, conn, peerConnection)
    me.Activate()

    peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
        go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}})
				if errSend != nil {
					fmt.Println(errSend)
				}
			}
		}()

        if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio{
            me.SetAudioTrack(remoteTrack)
        }else{
            me.SetVideoTrack(remoteTrack)
        }
    })

    // Set the remote SessionDescription
    err = peerConnection.SetRemoteDescription(offer)
    if err != nil {
        log.Fatalln(err)
    }

    // Create answer
    answer, err := peerConnection.CreateAnswer(nil)
    if err != nil {
        log.Fatalln(err)
    }

    // Sets the LocalDescription, and starts our UDP listeners
    err = peerConnection.SetLocalDescription(answer)
    if err != nil {
        log.Fatalln(err)
    }

    //Send SDP Answer
    respBody := constant.SDPResponse{}
    respBody.Action = "SERVER_ANSWER"
    respBody.SDP = answer
    ans, _ := json.Marshal(respBody)
    log.Println("[ACTION - INIT] - SDP Answer Sent")
    conn.WriteMessage(websocket.TextMessage, ans)

    //Loop over other clients in the room and consume tracks
    log.Println("[ACTION - INIT] - ROOM LENGTH", len(me.Room.Clients))
    time.Sleep(3 * time.Second) 
    if len(me.Room.Clients) > 1{
    	for he, status := range me.Room.Clients {
			if status {
				if he.UserID != me.UserID{

		            //Send SDP Answer
		            reqBody := constant.RequestBody{}
		            reqBody.Action = "RENEGOTIATE_EXIST_CLIENT"
		            reqBody.UserID = me.UserID
		            he.Sensor <- reqBody

				}else{
					//Send SDP Answer
		            reqBody := constant.RequestBody{}
		            reqBody.Action = "RENEGOTIATE_SELF_CLIENT"
		            reqBody.UserID = me.UserID
		            me.Sensor <- reqBody
				}
			}
		}
    }
	


}

func RespondToClientAnswer(reqBody constant.RequestBody){
	fmt.Println("***************************************************(   RESPOND TO CLIENT ANSWER    )*************************************")

	userID := reqBody.UserID
	answer := reqBody.SDP
	for client, status := range room.Clients {
        if status {
            //skip my tracks
            if client.UserID == userID{

                // Sets the LocalDescription, and starts our UDP listeners
                err := client.PC.SetRemoteDescription(answer)
                if err != nil {
                    log.Fatalln(err)
                }

                break;

            }
        }
    }
}


