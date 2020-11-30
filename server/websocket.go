package server

import (
    log "github.com/sirupsen/logrus"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "net/http"
    "github.com/voskos/voskos-rtc-sfu/parser"
    "github.com/voskos/voskos-rtc-sfu/router"

)

func CreteWebsocketServer(router *router.Router) {

    log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	gin.SetMode(gin.ReleaseMode)

    r := gin.Default()

    r.GET("/ws", func(c *gin.Context) {
        wshandler(router, c.Writer, c.Request)
    })

    r.Run("localhost:8080")
}

var wsupgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}



func wshandler(router *router.Router, w http.ResponseWriter, r *http.Request) {

    log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)

	wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
    conn, err := wsupgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Error("Failed to set websocket upgrade: %+v", err)
        return
    }

    for {

        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }

		log.Info("Web Socket received new message, sent for Parser")
        parser.ParseMessage(router, conn, msg)

    }
}