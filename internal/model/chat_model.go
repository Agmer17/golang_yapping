package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatModel struct {
	Id         uuid.UUID
	SenderId   uuid.UUID
	ReceiverId uuid.UUID
	ReplyTo    *uuid.UUID
	ChatText   *string
	PostId     *uuid.UUID
	IsRead     bool
	CreatedAt  time.Time
	IsOwn      bool
	Attachment []ChatAttachment
}

type ChatAttachment struct {
	Id        uuid.UUID `json:"id"`
	ChatId    uuid.UUID `json:"chat_id"`
	FileName  string    `json:"file_name"`
	MediaType string    `json:"media_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}
