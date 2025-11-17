package main

import (
	"github.com/Agmer17/golang_yapping/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {

	server := gin.Default()

	api := server.Group("/api")

	handlers.NewAuthHandler().RegisterRoutes(api)

	server.Run(":80")

}
