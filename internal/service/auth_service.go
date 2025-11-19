package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/Agmer17/golang_yapping/pkg/customerrors"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type ResponseSchema map[string]any

type AuthServiceInterface interface {
	LoginService(string, string, context.Context) (ResponseSchema, *customerrors.ServiceErrors)
	SignUp(username string, email string, fullName string, password string, c context.Context) (ResponseSchema, *customerrors.ServiceErrors)
}

type AuthService struct {
	UserRepo    repository.UserRepositoryInterface
	RedisClient *redis.Client
}

func NewAuthService(repo *repository.UserRepository, redclient *redis.Client) *AuthService {
	return &AuthService{
		UserRepo:    repo,
		RedisClient: redclient,
	}
}

func (a *AuthService) setRedisSession(
	uId string,
	refreshToken string,
	ctx context.Context,
) error {

	key := "session:" + refreshToken

	refreshClaims, err := pkg.VerifyRefreshToken(refreshToken)

	if err != nil {
		return err
	}

	expiresAt := refreshClaims.ExpiresAt.Time
	ttl := time.Until(expiresAt)

	data := map[string]any{
		"user_id":       uId,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt.Format(time.RFC3339),
		"created_at":    time.Now().Format(time.RFC3339),
	}

	// set hashmap
	if err := a.RedisClient.HSet(ctx, key, data).Err(); err != nil {
		return err
	}

	// set TTL Redis SESUAI exp refresh token
	if err := a.RedisClient.Expire(ctx, key, ttl).Err(); err != nil {
		return err
	}

	return nil
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
			Message: "ada kesalahan di server error : " + err.Error(),
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

	err = a.setRedisSession(data.Id.String(), refreshToken, ctx)

	if err != nil {
		return nil, &customerrors.ServiceErrors{
			Code:    http.StatusInternalServerError,
			Message: "Terjadi kesalahan di server : " + err.Error(),
		}
	}

	return ResponseSchema{"message": "berhasil login", "accessToken": accessToken, "refreshToken": refreshToken, "id": data.Id}, nil

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
