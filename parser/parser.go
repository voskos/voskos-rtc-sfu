package parser

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/voskos/voskos-rtc-sfu/action"
	"github.com/voskos/voskos-rtc-sfu/constant"
	"github.com/voskos/voskos-rtc-sfu/router"
	log "github.com/sirupsen/logrus"
)

func ParseMessage(router *router.Router, conn *websocket.Conn, msg []byte) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	
	reqBody := constant.RequestBody{}
	json.Unmarshal(msg, &reqBody)

	action_type := reqBody.Action
	log.Info("Message recieved with action : ", action_type)

	switch action_type {

	case "INIT":
		go action.Init(router, conn, reqBody)

	case "CLIENT_ANSWER":
		go action.RespondToClientAnswer(router, reqBody)

		// case "NEW_ICE_CANDIDATE_CLIENT":
		//     action.AddIceCandidate(router, reqBody)
	}
}
