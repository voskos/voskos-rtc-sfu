package main

import (
	"fmt"
	nats "github.com/nats-io/nats.go"

	"github.com/gin-gonic/gin"
	//"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
)

type message struct {
	Id string
	Content string
}

const NATS_HOST_ = "nats://ec2-3-7-252-87.ap-south-1.compute.amazonaws.com:4222"

func main() {
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

	//***************************************************************************************************//
	userIdMediaTrackMap := make(map[string]chan *webrtc.Track)
	fmt.Println(userIdMediaTrackMap)

	m := webrtc.MediaEngine{}
	// Setup the codecs you want to use.
	// Only support VP8(video compression), this makes our proxying code simpler
	m.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))
	fmt.Println(api)
	//


	//Starting gin server
	router := gin.Default()

	router.POST("/offer", func(c *gin.Context) {

	})

	router.Run(":8080")
}
