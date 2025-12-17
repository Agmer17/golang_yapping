package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatModel struct {
	Id         uuid.UUID
	SenderId   uuid.UUID
	ReceiverId uuid.UUID
	ReplyTo    uuid.UUID
	ChatText   *string
	ChatMedia  *string
	PostId     uuid.UUID
	IsRead     bool
	CreatedAt  time.Time
}
