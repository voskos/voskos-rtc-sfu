package util

import (
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/voskos/voskos-rtc-sfu/constant"
)

func SendErrMessage(msg string, conn *websocket.Conn){
    //Send SDP Answer
    respBody := constant.RequestBody{}
    respBody.Action = "ERR_OCCURED"
    respBody.Message = msg
    r, _ := json.Marshal(respBody)
    log.Info("Error Message", msg)
    conn.WriteMessage(websocket.TextMessage, r)
}