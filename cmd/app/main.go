package main

import (
	"context"
	"os"

	"github.com/Agmer17/golang_yapping/configs"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("gagal baca file env : " + err.Error())
	}

	dbUrl := os.Getenv("DATABASE_URL")
	redisUrl := os.Getenv("REDIS_URL")
	jwtSecret := os.Getenv("DATABASE_URL")

	pkg.JwtInit(jwtSecret)

	ctx := context.Background()
	redCtx := context.Background()

	app := configs.NewApp(ctx, dbUrl, redCtx, redisUrl)

	defer app.Shutdown()

	app.Run()
}
