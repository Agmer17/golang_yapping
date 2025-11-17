package configs

import (
	"github.com/Agmer17/golang_yapping/internal/handlers"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetUpRouter(p *pgxpool.Pool) *gin.Engine {

	userRepo := repository.NewUserRepo(p)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	server := gin.Default()

	api := server.Group("/api")
	authHandler.RegisterRoutes(api)

	return server

}
