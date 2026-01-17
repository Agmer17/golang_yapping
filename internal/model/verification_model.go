package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypePasswordVerification = "PASSWORD_RESET"
	TypeAccountVerification  = "ACCOUNT_VERIFICATION"
)

type VerificationModel struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Token     string
	Type      string
	UsedAt    time.Time
	CreatedAt time.Time
}
