package main

import (
	"github.com/voskos/voskos-rtc-sfu/router"
	"github.com/voskos/voskos-rtc-sfu/server"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	log.Info("Namaskar, Voskos! ", "Media server starting ...")

	router := router.NewRouter()
	go router.Run()
	
	server.CreteWebsocketServer(router)
}
