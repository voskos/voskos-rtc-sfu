package parser

import (
    "log"
    "encoding/json"
    "github.com/gorilla/websocket"
    "github.com/voskos/voskos-rtc-sfu/constant"
    "github.com/voskos/voskos-rtc-sfu/action"
    "github.com/voskos/voskos-rtc-sfu/router"
)

func ParseMessage(router *router.Router, conn *websocket.Conn, msg []byte){
    reqBody := constant.RequestBody{}
    json.Unmarshal(msg, &reqBody)

    action_type := reqBody.Action
    log.Println("[PARSER] - Message recieved with action : ", action_type)

    switch action_type {

    case "INIT":
        action.Init(router, conn, reqBody)

    case "CLIENT_ANSWER":
        action.RespondToClientAnswer(router, reqBody)

    // case "NEW_ICE_CANDIDATE_CLIENT":
    //     action.AddIceCandidate(router, reqBody)
    }
}