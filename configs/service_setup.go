package configs

import (
	"context"

	"github.com/Agmer17/golang_yapping/internal/event"
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

	EmailService *service.EmailService

	// event
	EventBus *event.EventBus
}

func NewServiceConfigs(
	pool *pgxpool.Pool,
	r *redis.Client,
	email string,
	emailPw string,
	eventContext context.Context,
) *serviceConfigs {
	hub := ws.NewHub()

	userRepo := repository.NewUserRepo(pool)
	chatRepo := repository.NewChatRepo(pool)
	chatAttachmentRepo := repository.NewChatAttachmentRepo(pool)

	authService := service.NewAuthService(userRepo, r)
	userService := service.NewUserService(userRepo)

	fileService := service.NewFileService()
	chatService := service.NewChatService(chatRepo, hub, userService, chatAttachmentRepo, fileService, r)

	emailService := service.NewEmailService(email, emailPw)

	// fmt.Println("======================================================")
	// fmt.Println(email)
	// fmt.Println(emailPw)
	// fmt.Println("======================================================")

	eventBus := event.NewEventBus(hub, eventContext)

	// setup event bus disini!
	event.SetupEvent(eventBus)

	return &serviceConfigs{
		AuthService:  authService,
		ChatService:  chatService,
		FileService:  fileService,
		UserService:  userService,
		Hub:          hub,
		EmailService: emailService,
		EventBus:     eventBus,
	}

}
