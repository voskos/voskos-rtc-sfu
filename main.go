package main

import (
	"log"
    "github.com/voskos/voskos-rtc-sfu/signal"
    "github.com/voskos/voskos-rtc-sfu/router"
)

func main() {
	room := router.NewRoom()
	go room.Run()
	log.Println("New room created")
    signal.CreteWebsocketSDPServer(room)
}