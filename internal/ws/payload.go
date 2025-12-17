package ws

import (
	"encoding/json"
)

const (
	ActionNotification   = "SYSTEM_NOTIFICATION"
	ActionSystem         = "SYSTEM"
	ActionSubscribe      = "SUBSCRIBE"
	ActionPrivateMessage = "PRIVATE_MESSAGE"

	TypeSystemError = "ERROR"
	TypeSystemOk    = "OK"
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

// user metadata

type UserMetadata struct {
	Username       string
	FullName       string
	ProfilePicture *string
}

type PrivateMessageData struct {
	Message  *string      `json:"text_message" binding:"required"`
	MediaUrl *string      `json:"media_url"`
	From     UserMetadata `json:"from"`
}
