package repository

import (
	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerificationRepoInterface interface {
	SetVerificationCode(payload model.VerificationModel) error
	FindVerificationCode(token string, verifType string) (model.VerificationModel, error)
}

type VerificationRepo struct {
	Pool *pgxpool.Pool
}

func NewVerificationRepo(pool *pgxpool.Pool) *VerificationRepo {

	return &VerificationRepo{
		Pool: pool,
	}
}

func (vr *VerificationRepo) SetVerificationCode(payload model.VerificationModel) error {

	return nil
}

func (vr *VerificationRepo) FindVerificationCode(token string, verifType string) (model.VerificationModel, error) {

	return model.VerificationModel{}, nil
}
