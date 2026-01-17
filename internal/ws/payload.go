package ws

import (
	"encoding/json"

	"github.com/google/uuid"
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
	Action   string          `json:"action"`
	Detail   string          `json:"detail"`
	Type     string          `json:"type"`
	Receiver string          `json:"-"`
	Data     json.RawMessage `json:"data"`
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
	Id             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	FullName       string    `json:"full_name"`
	ProfilePicture *string   `json:"profile_picture"`
}

type PrivateMessageData struct {
	Message  *string      `json:"text_message" binding:"required"`
	MediaUrl []string     `json:"media_url"`
	From     UserMetadata `json:"from"`
}
