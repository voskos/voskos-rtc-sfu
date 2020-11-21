package main

import (
	"log"
    "github.com/voskos/voskos-rtc-sfu/server"
    "github.com/voskos/voskos-rtc-sfu/router"
)

func main() {
	router := router.NewRouter()
	go router.Run()
	log.Println("[MAIN] - Server Initiated")
    server.CreteWebsocketServer(router)
}