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

    // Allow us to receive 1 video track
    if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeVideo); err != nil {
        log.Fatalln(err)
    }
    // Allow us to receive 1 audio track
    if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeAudio); err != nil {
        log.Fatalln(err)
    }

    me := router.AddClientToRoom(room, userID, conn, peerConnection)
    me.Activate()

    peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
        go func() {
            ticker := time.NewTicker(rtcpPLIInterval)
            for range ticker.C {
                if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
                    log.Println(rtcpSendErr)
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
    fmt.Println("ROOM LENGHT", len(me.Room.Clients))
    if len(me.Room.Clients) > 1{
    	for he, status := range me.Room.Clients {
			if status {
				fmt.Println(he.UserID, "*********************")
				if he.UserID != me.UserID{

		            //Send SDP Answer
		            reqBody := constant.RequestBody{}
		            reqBody.Action = "RENEGOTIATE_EXIST_CLIENT"
		            reqBody.UserID = me.UserID
		            he.Sensor <- reqBody

				}else{
					//Send SDP Answer
		            reqBody := constant.RequestBody{}
		            reqBody.Action = "RenegotiateDueToSelfJoin"
		            reqBody.UserID = me.UserID
		            me.Sensor <- reqBody
				}
			}
		}
    }
	


}


