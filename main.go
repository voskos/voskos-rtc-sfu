package main

import (
	"fmt"
	"log"
	"time"
	nats "github.com/nats-io/nats.go"

	//"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/voskos/voskos-rtc-sfu/signal"
)

type message struct {
	Id string
	Content string
}

const NATS_HOST_ = "nats://ec2-3-7-252-87.ap-south-1.compute.amazonaws.com:4222"
const rtcpPLIInterval = time.Second * 3


func main() {
	//****************************************SDP SERVER - BEGIN *************************************//

	// channels for send/recv SDPs
	sdpInChan, sdpOutChan := signal.HTTPSDPServer()

	offer := webrtc.SessionDescription{}
	signal.Decode(<-sdpInChan, &offer)
	log.Println("OFFER\n", offer) // json format

	// Since we are answering use PayloadTypes declared by offerer
	mediaEngine := webrtc.MediaEngine{}
	err := mediaEngine.PopulateFromSDP(offer)
	if err != nil {
		panic(err)
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

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
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		log.Fatalln(err)
	}

	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		log.Printf("Track recived of kind %s", remoteTrack.Kind())
		peerConnection.AddTrack(remoteTrack)

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

	// Create channel that is blocked until ICE Gathering is complete
	// gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	// <-gatherComplete

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("ANSWER for Caster => \n", answer) // json format of SDP Answer
	// Get the LocalDescription and take it to base64 so we can paste in browser
	baseAnswer := signal.Encode(answer)
	sdpOutChan <- baseAnswer
	//log.Println(baseAnswer)

	//****************************************SDP SERVER - END *************************************//


	//****************************************NATS SERVER - BEGIN ************************************//
	fmt.Printf("Connectivity checking with Nats server !!!\n")

	nc, _ := nats.Connect(NATS_HOST_)
	ec, _ := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	defer ec.Close()

	consumerCh := make(chan *message)
	ec.BindRecvChan("foo", consumerCh)

	producerCh := make(chan *message)
	ec.BindSendChan("foo", producerCh)

	msg := &message{Id: "Local System 001", Content: "Hello, World!"}
	producerCh <- msg
	producerCh <- msg

	who := <- consumerCh
	fmt.Printf("%+v\n",who)
	fmt.Printf("Connectivity Check completed successfully!!\n")

	//****************************************NATS SERVER - END  ************************************//
	for{

	}


}
