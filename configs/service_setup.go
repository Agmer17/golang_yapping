package configs

import (
	"context"

	"github.com/Agmer17/golang_yapping/internal/event"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type serviceConfigs struct {
	AuthService         *service.AuthService
	ChatService         *service.ChatService
	UserService         *service.UserService
	FileService         *service.FileStorage
	Hub                 *ws.Hub
	VerificationService *service.VerificationService

	EmailService *pkg.MailSender

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
	VerifcationRepo := repository.NewVerificationRepo(pool)

	// email sender
	emailService, err := pkg.NewMailSender(email, emailPw)

	// harus panik soalnya pasti salah konfigurasi antara di email atau app password
	if err != nil {
		panic(err)
	}

	// event bus backgorund job
	eventBus := event.NewEventBus(hub, eventContext, emailService)

	// setup event bus disini!
	event.SetupEvent(eventBus)

	authService := service.NewAuthService(userRepo, r, eventBus)
	userService := service.NewUserService(userRepo)

	fileService := service.NewFileService()
	chatService := service.NewChatService(chatRepo, hub, userService, chatAttachmentRepo, fileService, r, eventBus)
	verificationService := service.NewVerificationService(VerifcationRepo, r)

	return &serviceConfigs{
		AuthService:         authService,
		ChatService:         chatService,
		FileService:         fileService,
		UserService:         userService,
		Hub:                 hub,
		EmailService:        emailService,
		EventBus:            eventBus,
		VerificationService: verificationService,
	}

}
