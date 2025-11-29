package ws

import "encoding/json"

const (
	actionNotification = "SYSTEM_NOTIFICATION"
	actionSystem       = "SYSTEM"
	actionSubscribe    = "SUBSCRIBE"

	typeSystemError = "ERROR"
	typeSystemOk    = "OK"
)

type WebsocketEvent struct {
	Action string          `json:"action"`
	Detail string          `json:"detail"`
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data"`
}

type NotificationEventData struct {
	Type    string `json:"notification_type"`
	Message string `json:"notification_message"`
}

type JoinRoomEventData struct {
	JoinTo string `json:"join_to" binding:"required"`
}
