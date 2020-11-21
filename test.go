package main

import (
	"flag"
	"net/url"
	"encoding/json"
	"log"

	"github.com/pion/webrtc/v3"
	"github.com/gorilla/websocket"
	"github.com/voskos/voskos-rtc-sfu/constant"
)

var addr = flag.String("addr", "localhost:8080", "http service address")


func newPeerConnection() {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	// defer c.Close()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			_, message, _ := c.ReadMessage()
			//fmt.Printf("Type: %T, Message: %v", err, err)
			reqBody := constant.RequestBody{}
		    json.Unmarshal(message, &reqBody)

		    action_type := reqBody.Action
		    log.Println("[TEST] - Message recieved with action : ", action_type)

		    switch action_type {

		    case "SERVER_ANSWER":
		        peerConnection.SetRemoteDescription(reqBody.SDP)

		    case "SERVER_OFFER":
		        peerConnection.SetRemoteDescription(reqBody.SDP)
		        ans, _ := peerConnection.CreateAnswer(nil)
		        peerConnection.SetLocalDescription(ans)

		        reqBody := constant.RequestBody{}
			    reqBody.Action = "CLIENT_ANSWER"
			    reqBody.SDP = ans
			    reqBody.RoomID = "Room - 1"
			    reqBody.UserID = "User-001"
			    off, _ := json.Marshal(reqBody)
			    log.Println("[TEST] - SDP Answer Sent")
			    c.WriteMessage(websocket.TextMessage, off)

		    // case "NEW_ICE_CANDIDATE_CLIENT":
		    //     action.AddIceCandidate(router, reqBody)
		    }
		}
	}()


	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	gatherCompletePromise := webrtc.GatheringCompletePromise(peerConnection)

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	} else if err = peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}
	<-gatherCompletePromise


	//Send SDP Answer
    reqBody := constant.RequestBody{}
    reqBody.Action = "INIT"
    reqBody.SDP = offer
    reqBody.RoomID = "Room - 1"
    reqBody.UserID = "User-001"
    off, _ := json.Marshal(reqBody)
    log.Println("[TEST] - SDP Offer Sent")
    c.WriteMessage(websocket.TextMessage, off)


	// resp, err := http.Post(fmt.Sprintf("http://%s/doSignaling", os.Args[1]), "application/json", bytes.NewReader(offerJSON))
	// if err != nil {
	// 	panic(err)
	// }
	// resp.Close = true



	// var answer webrtc.SessionDescription
	// if err = json.NewDecoder(resp.Body).Decode(&answer); err != nil {
	// 	panic(err)
	// }

	// if err = peerConnection.SetRemoteDescription(answer); err != nil {
	// 	panic(err)
	// }
	// resp.Body.Close()
	for{}
}

func main() {

		for i := 0; i < 1; i++ {
			newPeerConnection()
		}
}
