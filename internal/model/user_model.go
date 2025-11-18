package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id             uuid.UUID
	FullName       string
	Username       string
	Email          string
	Password       string
	Role           string
	Birthday       *time.Time
	Bio            *string
	ProfilePicture *string
	BannerPicture  *string
	CreatedAt      time.Time
}
