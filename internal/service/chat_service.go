package service

import (
	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/google/uuid"
)

type ChatServiceInterface interface {
	SaveChat(model.ChatModel) error
	GetChatBeetween(r uuid.UUID, s uuid.UUID) []model.ChatModel
}

type ChatService struct {
	Pool repository.ChatRepositoryInterface
}

func NewChatService(c repository.ChatRepositoryInterface) *ChatService {
	return &ChatService{
		Pool: c,
	}
}

func (cs *ChatService) SaveChat(d model.ChatModel) error {

	return nil

}

func (cs *ChatService) GetChatBeetween(sender uuid.UUID, receiver uuid.UUID) []model.ChatModel {

	return nil

}
