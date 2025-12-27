package configs

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	DB      *pgxpool.Pool
	Router  *gin.Engine
	Redis   *redis.Client
	Service *serviceConfigs
}

func NewApp(ctx context.Context, dbUrl string, redCtx context.Context, redUrl string) *App {

	pool, err := SetUpDatabase(ctx, dbUrl)

	rdb := SetUpRedis(redCtx, redUrl)

	if err != nil {
		panic(err)
	}

	svc := NewServiceConfigs(pool, rdb)
	r := SetUpRouter(pool, rdb, svc)

	return &App{
		DB:      pool,
		Router:  r,
		Redis:   rdb,
		Service: svc,
	}

}

func (a *App) Run() error {
	return a.Router.Run(":80")
}

func (a *App) Shutdown() {
	a.DB.Close()
}
