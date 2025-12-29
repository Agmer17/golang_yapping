package pkg

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

func IsPStrEmpty(s *string) bool {
	return s == nil || *s == ""
}

func StringToUuid(e *string) (*uuid.UUID, error) {
	if !IsPStrEmpty(e) {
		rep, err := uuid.Parse(*e)
		if err != nil {
			return nil, err
		}

		return &rep, nil
	}

	return nil, nil
}

func GenerateRandomStringToken(n int) (string, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
