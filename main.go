package main

import (
	"log"
    "github.com/voskos/voskos-rtc-sfu/server"
)

func main() {
	log.Println("[MAIN] - Server Initiated")
    server.CreteWebsocketServer()
}