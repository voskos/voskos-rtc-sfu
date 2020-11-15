package signal

import (
    "fmt"
    "log"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "encoding/json"
    "net/http"
    "github.com/pion/rtcp"
    "github.com/pion/webrtc/v3"
    "github.com/voskos/voskos-rtc-sfu/router"

)

func CreteWebsocketSDPServer(room *router.Room) {

    r := gin.Default()

    r.GET("/ws", func(c *gin.Context) {
        wshandler(room, c.Writer, c.Request)
    })

    r.Run("localhost:8080")
}

var wsupgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type SDPBody struct {
    Action string `json:"action"`
    UserID string `json:"user_id"`
    SDP webrtc.SessionDescription `json:"sdp"`
}

const (
    // PLI (Pictire Loss Indication)
    rtcpPLIInterval = time.Second * 3
)

func wshandler(room *router.Room, w http.ResponseWriter, r *http.Request) {
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
    conn, err := wsupgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }

    for {

        t, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }

        //Recieve offer from client
        reqBody := SDPBody{}
        json.Unmarshal(msg, &reqBody)

        action := reqBody.Action
        uid := reqBody.UserID
        offer := reqBody.SDP

        //**************************************** CLIENT OFFER ACTION ****************************//
        if action == "CLIENT_OFFER" {
            log.Println("New connection from UID ------", uid)

            //create a peerconnection and client object
            // m := webrtc.MediaEngine{}
            // // Setup the codecs you want to use.
            // // Only support VP8, this makes our proxying code simpler
            // // m.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
            // // m.RegisterCodec(webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000))

            // // Create the API object with the MediaEngine
            // api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

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


            client := router.AddClientToRoom(room, uid, conn, peerConnection)
            go client.ConsumeAudioTracks(peerConnection)
            go client.ConsumeVideoTracks(peerConnection)

            // // Allow us to receive 1 video track
            // if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeVideo); err != nil {
            //     log.Fatalln(err)
            // }
            // // Allow us to receive 1 audio track
            // if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeAudio); err != nil {
            //     log.Fatalln(err)
            // }

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
                    client.SetAudioTrack(remoteTrack)
                    log.Println("Audio track saved for ", uid)
                }else{
                    client.SetVideoTrack(remoteTrack)
                    log.Println("Video track saved for ", uid)
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
            respBody := SDPBody{}
            respBody.Action = "SERVER_ANSWER"
            respBody.UserID = uid
            respBody.SDP = answer
            ans, _ := json.Marshal(respBody)
            log.Println("SDP Answer Sent")
            conn.WriteMessage(t, ans)
        }

    }
}