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
	GetUserDataByUsername(string, context.Context) (*model.User, error)
	AddUser(model.User, context.Context) (model.User, error)
	DeleteUser(uuid.UUID, context.Context) error
	EditUser(model.User, context.Context) (model.User, error)
	ExistByNameOrUsername(username string, email string, c context.Context) (bool, error)
}

type UserRepository struct {
	Pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		Pool: pool,
	}
}

func (u *UserRepository) GetUserDataByUsername(username string, c context.Context) (*model.User, error) {

	q := `
		select id, 
			full_name, 
			username, 
			email, 
			role,
			password, 
			birthday, 
			bio, 
			profile_picture, 
			banner_picture, 
			created_at
		from users
		where username = $1
		limit 1
		`

	var user model.User

	err := u.Pool.QueryRow(c, q, username).Scan(
		&user.Id,
		&user.FullName,
		&user.Username,
		&user.Email,
		&user.Role,
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

func (u *UserRepository) DeleteUser(id uuid.UUID, ctx context.Context) error {
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

func (u *UserRepository) ExistByNameOrUsername(username string, email string, c context.Context) (bool, error) {
	var exist bool

	q := `
        SELECT EXISTS (
            SELECT 1
            FROM users u
            WHERE u.username = $1 OR u.email = $2
        );
    `

	err := u.Pool.QueryRow(c, q, username, email).Scan(&exist)
	if err != nil {
		return false, err
	}

	return exist, nil
}

func (u *UserRepository) EditUser(e model.User, c context.Context) (model.User, error) {

	// todo: IMPLEMENT INI
	return model.User{}, nil

}
