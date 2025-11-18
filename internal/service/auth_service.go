package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/Agmer17/golang_yapping/pkg/customerrors"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type ResponseSchema map[string]any

type AuthServiceInterface interface {
	LoginService(string, string, context.Context) (ResponseSchema, *customerrors.ServiceErrors)
	SignUp(username string, email string, fullName string, password string, c context.Context) (ResponseSchema, *customerrors.ServiceErrors)
}

type AuthService struct {
	UserRepo repository.UserRepositoryInterface
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{
		UserRepo: repo,
	}
}

func (a *AuthService) LoginService(username string, pw string, ctx context.Context) (ResponseSchema, *customerrors.ServiceErrors) {
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

	accessToken, err := pkg.GenerateToken(data.Id, data.Role, 10)

	if err != nil {
		return nil, &customerrors.ServiceErrors{
			Code:    http.StatusInternalServerError,
			Message: "Terjadi kesalahan di server : " + err.Error(),
		}
	}

	refreshToken, err := pkg.GenerateTokenNoRole(data.Id, 10800)

	return ResponseSchema{"message": "berhasil login", "accessToken": accessToken, "refreshToken": refreshToken}, nil

}

func (a *AuthService) SignUp(username string, email string, fullName string, password string, c context.Context) (ResponseSchema, *customerrors.ServiceErrors) {

	var newUser model.User

	ex, err := a.UserRepo.ExistByNameOrUsername(username, email, c)
	if ex {
		return nil, &customerrors.ServiceErrors{
			Code:    http.StatusConflict,
			Message: "username atau email sudah terdaftar",
		}
	}

	newUser.Username = username
	newUser.Email = email

	newUser.FullName = fullName
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, &customerrors.ServiceErrors{
			Code:    500,
			Message: "Terjadi kesalahan di server silahkan coba lagi nanti",
		}
	}

	newUser.Password = string(hashedPw)

	data, err := a.UserRepo.AddUser(newUser, c)

	if err != nil {
		return nil, &customerrors.ServiceErrors{
			Code:    500,
			Message: err.Error(),
		}
	}

	return ResponseSchema{
		"message":    "berhasil membuat akun",
		"created_at": data.CreatedAt,
	}, nil
}
