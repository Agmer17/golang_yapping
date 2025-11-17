package repository

import (
	"context"
	"errors"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryInterface interface {
	GetUserData(string, context.Context) (*model.User, error)
	AddUser(model.User) model.User
	DeleteUser(uuid.UUID, context.Context) error
	EditUser(model.User, context.Context) (model.User, error)
}

type UserRepository struct {
	Pool *pgxpool.Pool
}

func (u *UserRepository) GetUserData(username string, c context.Context) (*model.User, error) {

	q := "select * from users where username = $1 limit 1"

	var user model.User

	err := u.Pool.QueryRow(c, q, username).Scan(
		&user.Id,
		&user.FullName,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Birthday,
		&user.Bio,
		&user.ProfilePicture,
		&user.BannerPicture,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *UserRepository) AddUser(user model.User, c context.Context) (model.User, error) {
	// ini minta field wajib
	// username, email, fullname. password
	var nu model.User

	err := pgx.BeginFunc(c, u.Pool, func(tx pgx.Tx) error {
		q := `
		insert into users (username, full_name, email, password)
		values($1, $2, $3, $4)
		returning id,username, full_name, email, password`

		return tx.QueryRow(c, q,
			user.Username,
			user.FullName,
			user.Email,
			user.Password,
		).Scan(
			&nu.Id,
			&nu.Username,
			&nu.FullName,
			&nu.Email,
			&nu.Password,
		)
	})

	if err != nil {
		return model.User{}, err
	}

	return nu, nil

}

func (u *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := pgx.BeginFunc(ctx, u.Pool, func(tx pgx.Tx) error {
		q := `delete from users where id = $1`

		tag, execErr := tx.Exec(ctx, q, id)
		if execErr != nil {
			return execErr
		}

		if tag.RowsAffected() == 0 {
			return errors.New("user tidak ditemukan!")
		}

		return nil
	})

	return err
}

func (u *UserRepository) EditUser(e model.User, c context.Context) (model.User, error) {

	// todo: IMPLEMENT INI
	return model.User{}, nil

}
