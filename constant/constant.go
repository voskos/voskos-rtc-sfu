package constant

import (
	"github.com/pion/webrtc/v3"
)

type RequestBody struct {
    Action string `json:"action"`
    UserID string `json:"user_id"`
    RoomID string `json:"room_id"`
    Body map[string]string `json:"body"`
    SDP webrtc.SessionDescription `json:"sdp"`
    ICE_Candidate *webrtc.ICECandidate `json:"ice_candidate"`

}

type SDPResponse struct{
	Action string `json:"action"`
	SDP webrtc.SessionDescription `json:"sdp"`
}

type ICEResponse struct{
	Action string `json:"action"`
	ICE_Candidate *webrtc.ICECandidate `json:"ice_candidate"`
}