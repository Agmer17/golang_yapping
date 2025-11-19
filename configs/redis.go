package configs

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func SetUpRedis(ctx context.Context, addr string) *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0, // default db ini ya
	})

	return rdb
}
