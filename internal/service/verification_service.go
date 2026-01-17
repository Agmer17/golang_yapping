package service

import (
	"context"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type VerificationServiceInterface interface {
	SaveVerificationCode(ctx context.Context, userId uuid.UUID, tokenType string) error
	GetVerificationCode(ctx context.Context, token string) (model.VerificationModel, error)
}

type VerificationService struct {
	VerifRepo repository.VerificationRepoInterface
	RedisCli  *redis.Client
}

func NewVerificationService(repo *repository.VerificationRepo, redCli *redis.Client) *VerificationService {

	return &VerificationService{
		VerifRepo: repo,
		RedisCli:  redCli,
	}
}

func (svc *VerificationService) SaveVerificationCode(ctx context.Context, userId uuid.UUID, tokenType string) error {

	return nil
}
func (svc *VerificationService) GetVerificationCode(ctx context.Context, token string) (model.VerificationModel, error) {

	return model.VerificationModel{}, nil
}
