package signal

import (
    "fmt"
    "log"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "encoding/json"
    "net/http"
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

type ClientOffer struct {
    UserID string `json:"user_id"`
    SDP webrtc.SessionDescription `json:"sdp"`
}


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
        reqBody := ClientOffer{}
        json.Unmarshal(msg, &reqBody)

        uid := reqBody.UserID
        offer := reqBody.SDP

        log.Println("UID ------ \n", uid)
        log.Println("OFFER ------\n", offer) // json format

        //create a peerconnection and client object
        m := webrtc.MediaEngine{}

        //util.DebugPrintf("here")
        // Setup the codecs you want to use.
        // Only support VP8, this makes our proxying code simpler
        m.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
        m.RegisterCodec(webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000))
        // Create the API object with the MediaEngine
        api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

        peerConnectionConfig := webrtc.Configuration{
            ICEServers: []webrtc.ICEServer{
                {
                    URLs: []string{"stun:stun.l.google.com:19302"},
                },
            },
        }

        // Create a new RTCPeerConnection
        peerConnection, err := api.NewPeerConnection(peerConnectionConfig)
        if err != nil {
            log.Fatalln(err)
        }

        // Allow us to receive 1 video track
        if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeVideo); err != nil {
            log.Fatalln(err)
        }

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

        ans, _ := json.Marshal(answer)
        log.Println("SDP Answer Sent")
        conn.WriteMessage(t, ans)

        router.AddClientToRoom(room, uid, conn, peerConnection)




    }
}