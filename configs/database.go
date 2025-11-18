package configs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetUpDatabase(ctx context.Context, url string) (*pgxpool.Pool, error) {

	connStr := "postgres://postgres:agmer@localhost:5432/yapping_db?sslmode=disable"

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 20 * time.Minute
	config.MaxConnLifetime = 10 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil

}
