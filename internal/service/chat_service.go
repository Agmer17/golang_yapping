package service

import (
	"github.com/Agmer17/golang_yapping/internal/repository"
)

type ChatService struct {
	Pool repository.ChatRepositoryInterface
}

func NewChatService(c repository.ChatRepositoryInterface) *ChatService {
	return &ChatService{
		Pool: c,
	}

}
