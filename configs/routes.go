package configs

import (
	"github.com/Agmer17/golang_yapping/internal/handlers"
	"github.com/Agmer17/golang_yapping/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func SetUpRouter(p *pgxpool.Pool, r *redis.Client, svc *serviceConfigs) *gin.Engine {

	authHandler := handlers.NewAuthHandler(svc.AuthService)

	// ------------------- PROTECTED --------------------
	userHandler := handlers.NewUserHandler(svc.UserService)
	wsHandler := handlers.NewWebsocketHandler(svc.Hub)
	chatHandler := handlers.NewChatHandler(svc.ChatService)
	// --------------------------------------------------

	server := gin.Default()
	server.Use(cors.Default())

	api := server.Group("/api")
	api.Static("/uploads", "./uploads/public")

	authHandler.RegisterRoutes(api)

	protected := api.Group("/")

	protected.Use(middleware.AuthMiddleware())
	userHandler.RegisterRoutes(protected)
	wsHandler.RegisterRoutes(protected)
	chatHandler.RegisterRoutes(protected)

	return server

}
