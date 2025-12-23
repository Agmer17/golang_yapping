package configs

import (
	"github.com/Agmer17/golang_yapping/internal/handlers"
	"github.com/Agmer17/golang_yapping/internal/middleware"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func SetUpRouter(p *pgxpool.Pool, r *redis.Client) *gin.Engine {

	// hub for chatservice and ws
	hub := ws.NewHub()

	// ---------------- REPOSITORY ---------------
	userRepo := repository.NewUserRepo(p)
	chatRepo := repository.NewChatRepo(p)
	chatAttachmentRepo := repository.NewChatAttachmentRepo(p)
	// --------------------------------------------

	// unprotected
	authService := service.NewAuthService(userRepo, r)
	authHandler := handlers.NewAuthHandler(authService)

	// ------------------- service --------------------
	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(chatRepo, hub, userService, chatAttachmentRepo)

	// --------------------------------------------------

	// ------------------- PROTECTED --------------------

	userHandler := handlers.NewUserHandler(userService)
	wsHandler := handlers.NewWebsocketHandler(hub)
	chatHandler := handlers.NewChatHandler(chatService)
	// --------------------------------------------------

	server := gin.Default()
	server.Use(cors.Default())

	api := server.Group("/api")
	authHandler.RegisterRoutes(api)

	protected := api.Group("/")

	protected.Use(middleware.AuthMiddleware())
	userHandler.RegisterRoutes(protected)
	wsHandler.RegisterRoutes(protected)
	chatHandler.RegisterRoutes(protected)

	return server

}
