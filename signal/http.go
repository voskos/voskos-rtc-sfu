package signal

import (
	"fmt"
	"log"
	"flag"
	"io/ioutil"
	"net/http"
	"strconv"
)

//HTTPSDPServer starts a HTTP Server that consumes SDPs
//Creats a channel for recieving SDP and returns it
func HTTPSDPServer() (chan string, chan string) {
	port := flag.Int("port", 8080, "http server port")
	flag.Parse()

	sdpInChan := make(chan string)
	sdpOutChan := make(chan string)

	http.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("/sdp connected from %s", r.Host)
		body, _ := ioutil.ReadAll(r.Body)
		// process request of sdp
		sdpInChan <- string(body)
		// send response of sdp
		fmt.Fprintf(w, <-sdpOutChan)
		log.Println("sent base64 SDP to client")
	})

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(*port), nil)
		if err != nil {
			panic(err)
		}
	}()

	log.Println("WebRTC SFU example server is started")
	log.Printf("Started http SDP server on :%d", *port)
	return sdpInChan, sdpOutChan
}
