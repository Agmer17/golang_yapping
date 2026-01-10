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
	jwtSecret := os.Getenv("JWT_SECRET")

	pkg.JwtInit(jwtSecret)

	// email
	emailConfig := os.Getenv("EMAIL_CONFIG")
	emailPassword := os.Getenv("EMAIL_PASSWORD")

	ctx := context.Background()
	redCtx := context.Background()
	eventContext := context.Background()

	app := configs.NewApp(ctx, dbUrl, redCtx, redisUrl, eventContext, emailConfig, emailPassword)

	defer app.Shutdown()

	app.Run()
}
