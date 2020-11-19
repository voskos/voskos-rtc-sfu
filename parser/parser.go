package parser

import (
    "log"
    "encoding/json"
    "github.com/gorilla/websocket"
    "github.com/voskos/voskos-rtc-sfu/constant"
    "github.com/voskos/voskos-rtc-sfu/action"
)

func ParseMessage(conn *websocket.Conn, msg []byte){
    reqBody := constant.RequestBody{}
    json.Unmarshal(msg, &reqBody)

    action_type := reqBody.Action
    log.Println("[PARSER] - Message recieved with action : ", action_type)

    switch action_type {

    case "INIT":
        action.Init(conn, reqBody)

    case "CLIENT_ANSWER":
        action.RespondToClientAnswer(reqBody)
    }
}