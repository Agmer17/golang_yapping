package configs

import (
	"github.com/Agmer17/golang_yapping/internal/handlers"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func SetUpRouter(p *pgxpool.Pool, r *redis.Client) *gin.Engine {

	userRepo := repository.NewUserRepo(p)

	// unprotected
	authService := service.NewAuthService(userRepo, r)
	authHandler := handlers.NewAuthHandler(authService)

	// ------------------- PROTECTED --------------------

	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	// --------------------------------------------------

	// ------------------ WEBSOCKET ---------------------
	wsHandler := handlers.NewWebsocketHandler()
	// --------------------------------------------------

	server := gin.Default()
	server.Use(cors.Default())

	api := server.Group("/api")
	authHandler.RegisterRoutes(api)
	userHandler.RegisterRoutes(api)
	wsHandler.RegisterRoutes(api)

	return server

}
