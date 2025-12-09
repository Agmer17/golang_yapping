package ws

import (
	"encoding/json"

	"github.com/google/uuid"
)

const (
	actionNotification   = "SYSTEM_NOTIFICATION"
	actionSystem         = "SYSTEM"
	actionSubscribe      = "SUBSCRIBE"
	actionPrivateMessage = "PRIVATE_MESSAGE"

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

type PrivateMessageData struct {
	To        uuid.UUID `json:"to" binding:"required"`
	Message   string    `json:"text_message" binding:"required"`
	Media_url string    `json:"media_url"`
	From      uuid.UUID `json:"from"`
}
