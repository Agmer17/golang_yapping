package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/pkg/customerrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type publicUserData struct {
	Id             uuid.UUID  `json:"id"`
	Username       string     `json:"username"`
	FullName       string     `json:"full_name"`
	ProfilePicture *string    `json:"profile_picture"`
	BannerPicture  *string    `json:"banner_picture"`
	Bio            *string    `json:"bio"`
	Birthday       *time.Time `json:"birthday"`
	CreatedAt      time.Time  `json:"created_at"`
}

type UserServiceInterface interface {
	GetUserData(username string) (ResponseSchema, error)
	GetMyProfile(id uuid.UUID, ctx context.Context) (ResponseSchema, error)
}

type UserService struct {
	Pool repository.UserRepositoryInterface
}

func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{
		Pool: r,
	}
}

func (u *UserService) GetUserData(username string) (ResponseSchema, error) {
	return nil, nil
}

func (u *UserService) GetMyProfile(id uuid.UUID, ctx context.Context) (ResponseSchema, error) {

	myData, err := u.Pool.GetUserDataById(id, ctx)

	respData := publicUserData{
		Id:             myData.Id,
		Username:       myData.Username,
		FullName:       myData.FullName,
		ProfilePicture: myData.ProfilePicture,
		BannerPicture:  myData.BannerPicture,
		Bio:            myData.Bio,
		Birthday:       myData.Birthday,
		CreatedAt:      myData.CreatedAt,
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &customerrors.ServiceErrors{
				Code:    http.StatusNotFound,
				Message: "username tidak ditemukan",
			}
		}

		return nil, &customerrors.ServiceErrors{
			Code:    http.StatusInternalServerError,
			Message: "Terjadi kesalahan di server : " + err.Error(),
		}
	}

	return ResponseSchema{
		"data": respData,
	}, nil
}
