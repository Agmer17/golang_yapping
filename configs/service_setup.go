package configs

import (
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type serviceConfigs struct {
	AuthService *service.AuthService
	ChatService *service.ChatService
	UserService *service.UserService
	FileService *service.FileStorage
	Hub         *ws.Hub
}

func NewServiceConfigs(pool *pgxpool.Pool, r *redis.Client) *serviceConfigs {
	hub := ws.NewHub()

	userRepo := repository.NewUserRepo(pool)
	chatRepo := repository.NewChatRepo(pool)
	chatAttachmentRepo := repository.NewChatAttachmentRepo(pool)

	authService := service.NewAuthService(userRepo, r)
	userService := service.NewUserService(userRepo)

	fileService := service.NewFileService()
	chatService := service.NewChatService(chatRepo, hub, userService, chatAttachmentRepo, fileService)

	return &serviceConfigs{
		AuthService: authService,
		ChatService: chatService,
		FileService: fileService,
		UserService: userService,
		Hub:         hub,
	}

}
