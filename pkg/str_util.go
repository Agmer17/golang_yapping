package pkg

import (
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
