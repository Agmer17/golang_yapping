package service

import "github.com/Agmer17/golang_yapping/pkg/errors"

type AuthService interface {
	LoginService() (map[string]any, errors.ServiceErrors)
	SignUpService() (map[string]any, errors.ServiceErrors)
}
