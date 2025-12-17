package service

import (
	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/Agmer17/golang_yapping/pkg/customerrors"
	"github.com/google/uuid"
)

type ChatServiceInterface interface {
	SaveChat(*model.ChatModel) *customerrors.ServiceErrors
	GetChatBeetween(r uuid.UUID, s uuid.UUID) []model.ChatModel
}

type ChatService struct {
	Pool repository.ChatRepositoryInterface
	Hub  *ws.Hub
}

func NewChatService(c repository.ChatRepositoryInterface, h *ws.Hub) *ChatService {
	return &ChatService{
		Pool: c,
		Hub:  h,
	}
}

func (cs *ChatService) SaveChat(d *model.ChatModel) *customerrors.ServiceErrors {

	isValid := isChatValid(d)
	if !isValid {
		return &customerrors.ServiceErrors{
			Code:    400,
			Message: "Input tidak valid!",
		}

	}

	return nil

}

func (cs *ChatService) GetChatBeetween(sender uuid.UUID, receiver uuid.UUID) []model.ChatModel {

	return nil

}

func isPStrEmpty(s *string) bool {
	return s == nil || *s == ""
}

func isChatValid(d *model.ChatModel) bool {
	chatTextEmpty := isPStrEmpty(d.ChatText)
	chatMediaEmpty := isPStrEmpty(d.ChatMedia)
	postIdEmpty := d.PostId == uuid.Nil

	if chatTextEmpty && chatMediaEmpty && postIdEmpty {
		return false
	}

	return true
}

func (cs *ChatService) sendMessage(d model.ChatModel) {

}
