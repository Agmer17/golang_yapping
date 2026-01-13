package repository

import "github.com/Agmer17/golang_yapping/internal/model"

type VerificationRepoInterface interface {
	SetVerificationCode(payload model.VerificationModel) error
	FindVerificationCode(token string, verifType string) (model.VerificationModel, error)
}
