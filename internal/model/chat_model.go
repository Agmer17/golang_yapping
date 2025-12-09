package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ChatModel struct {
	Id         uuid.UUID
	SenderId   uuid.UUID
	ReceiverId uuid.UUID
	ReplyTo    uuid.UUID
	ChatText   sql.NullString
	ChatMedia  sql.NullString
	PostId     uuid.UUID
	IsRead     bool
	CreatedAt  time.Time
}
