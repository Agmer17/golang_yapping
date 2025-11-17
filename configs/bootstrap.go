package configs

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DB     *pgxpool.Pool
	Router *gin.Engine
}

func NewApp(ctx context.Context) *App {

	pool, err := SetUpDatabase(ctx)

	if err != nil {
		panic(err)
	}

	r := SetUpRouter(pool)

	return &App{
		DB:     pool,
		Router: r,
	}

}

func (a *App) Run() error {
	return a.Router.Run(":80")
}

func (a *App) Shutdown() {
	a.DB.Close()
}
