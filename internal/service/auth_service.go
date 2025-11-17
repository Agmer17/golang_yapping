package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/pkg/customerrors"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	LoginService(string, string, context.Context) (map[string]any, *customerrors.ServiceErrors)
}

type AuthService struct {
	UserRepo repository.UserRepositoryInterface
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{
		UserRepo: repo,
	}
}

func (a *AuthService) LoginService(username string, pw string, ctx context.Context) (map[string]any, *customerrors.ServiceErrors) {
	data, err := a.UserRepo.GetUserDataByUsername(username, ctx)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &customerrors.ServiceErrors{
				Code:    http.StatusNotFound,
				Message: "username tidak ditemukan",
			}
		}

		return nil, &customerrors.ServiceErrors{
			Code:    http.StatusInternalServerError,
			Message: "ada kesalahan di server",
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(pw))

	if err != nil {
		return nil, customerrors.New(http.StatusUnauthorized, "usename atau password salah")
	}

	return map[string]any{"message": "berhasil login", "data": data}, nil

}
