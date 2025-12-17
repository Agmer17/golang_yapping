package service

import (
	"context"
	"encoding/json"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/Agmer17/golang_yapping/pkg/customerrors"
	"github.com/google/uuid"
)

type ChatServiceInterface interface {
	SaveChat(m *model.ChatModel, ctx context.Context) *customerrors.ServiceErrors
	GetChatBeetween(r uuid.UUID, s uuid.UUID) []model.ChatModel
}

type ChatService struct {
	Pool repository.ChatRepositoryInterface
	Hub  *ws.Hub
	usv  *UserService
}

func NewChatService(c repository.ChatRepositoryInterface, h *ws.Hub, u *UserService) *ChatService {
	return &ChatService{
		Pool: c,
		Hub:  h,
		usv:  u,
	}
}

func (cs *ChatService) SaveChat(d *model.ChatModel, ctx context.Context) *customerrors.ServiceErrors {

	isValid := isChatValid(d)
	if !isValid {
		return &customerrors.ServiceErrors{
			Code:    400,
			Message: "Input tidak valid!",
		}

	}

	senderMetaData, err := cs.usv.GetUserDataById(d.SenderId, ctx)

	if err != nil {
		return &customerrors.ServiceErrors{
			Code:    404,
			Message: "sender tidak ditemukan!",
		}
	}

	go cs.sendMessage(d, senderMetaData)

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

func (cs *ChatService) sendMessage(d *model.ChatModel, senderMetaData publicUserData) {

	MessageData, _ := json.Marshal(ws.PrivateMessageData{
		Message:  d.ChatText,
		MediaUrl: d.ChatMedia,
		From: ws.UserMetadata{
			Username:       senderMetaData.Username,
			FullName:       senderMetaData.FullName,
			ProfilePicture: senderMetaData.ProfilePicture,
		},
	})

	event := ws.WebsocketEvent{
		Action: ws.ActionPrivateMessage,
		Detail: "Somebody send you a message",
		Type:   ws.TypeSystemOk,
		Data:   MessageData,
	}

	payload, _ := json.Marshal(event)
	rcvRoomId := "user:" + d.ReceiverId.String()

	cs.Hub.SendPayloadTo(rcvRoomId, payload)

}
