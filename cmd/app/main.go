package main

import (
	"context"

	"github.com/Agmer17/golang_yapping/configs"
)

func main() {
	ctx := context.Background()

	app := configs.NewApp(ctx)

	defer app.Shutdown()

	app.Run()
}
