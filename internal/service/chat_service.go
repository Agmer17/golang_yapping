package service

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/Agmer17/golang_yapping/pkg/customerrors"
	"github.com/google/uuid"
)

type ChatPostInput struct {
	SenderId   uuid.UUID
	ReceiverId string
	ReplyTo    *string
	ChatText   *string
	PostId     *string
	MediaFiles []*multipart.FileHeader
}

type ChatServiceInterface interface {
	SaveChat(m *ChatPostInput, ctx context.Context) *customerrors.ServiceErrors
	GetChatBeetween(r uuid.UUID, s uuid.UUID) []model.ChatModel
}

type ChatService struct {
	Pool    repository.ChatRepositoryInterface
	Hub     *ws.Hub
	usv     *UserService
	chatAtt repository.ChatAttachmentInterface
	storage *FileStorage
}

func NewChatService(c *repository.ChatRepository,
	h *ws.Hub,
	u *UserService,
	ct *repository.ChatAttachmentRepository,
	fileService *FileStorage) *ChatService {
	return &ChatService{
		Pool:    c,
		Hub:     h,
		usv:     u,
		chatAtt: ct,
		storage: fileService,
	}
}

func (cs *ChatService) SaveChat(d *ChatPostInput, ctx context.Context) *customerrors.ServiceErrors {

	isValid := isChatValid(d)
	if !isValid {
		return &customerrors.ServiceErrors{
			Code:    400,
			Message: "Input tidak valid!",
		}

	}

	cm, err := parseToChatModel(d)
	if err != nil {
		return &customerrors.ServiceErrors{
			Code:    500,
			Message: "Ada kesalahan saat parsing data " + err.Error(),
		}
	}

	savedChat, err := cs.Pool.Save(cm, ctx)
	if err != nil {
		return &customerrors.ServiceErrors{
			Code:    http.StatusInternalServerError,
			Message: "Ada kesalahan saat menyimpan pesan  " + err.Error(),
		}
	}

	if len(d.MediaFiles) != 0 {
		listMetadata, svcErr := cs.processAttachment(d.MediaFiles, savedChat.Id)
		if svcErr != nil {
			return svcErr
		}

		err := cs.chatAtt.SaveAll(listMetadata, ctx)
		if err != nil {
			cs.cleanUpAttachment(listMetadata)
			return &customerrors.ServiceErrors{
				Code:    http.StatusInternalServerError,
				Message: "Gagal saat menyimpan media " + err.Error(),
			}
		}

		go cs.sendChat(d, listMetadata)

		return nil
	}

	go cs.sendChat(d, []model.ChatAttachment{})
	return nil
}

func (cs *ChatService) GetChatBeetween(sender uuid.UUID, receiver uuid.UUID) []model.ChatModel {

	return nil

}

func isChatValid(d *ChatPostInput) bool {
	chatTextEmpty := pkg.IsPStrEmpty(d.ChatText)
	chatMediaEmpty := len(d.MediaFiles) == 0
	postIdEmpty := pkg.IsPStrEmpty(d.PostId)

	if chatTextEmpty && chatMediaEmpty && postIdEmpty {
		return false
	}

	return true
}

func (cs *ChatService) processAttachment(att []*multipart.FileHeader, chatId uuid.UUID) ([]model.ChatAttachment, *customerrors.ServiceErrors) {

	var chatAttachments []model.ChatAttachment = make([]model.ChatAttachment, 0)

	for _, v := range att {
		mimeType, err := cs.storage.DetectFileType(v)
		if err != nil {
			cs.cleanUpAttachment(chatAttachments)
			return nil, &customerrors.ServiceErrors{
				Code:    http.StatusInternalServerError,
				Message: "Terjadi kesalahan saat menyimpan file " + err.Error(),
			}
		}

		ext, ok := cs.storage.IsTypeSupportted(mimeType)

		if !ok {
			cs.cleanUpAttachment(chatAttachments)
			return nil, &customerrors.ServiceErrors{
				Code:    http.StatusBadRequest,
				Message: "File saat ini tidak didukung!",
			}

		}

		fName, err := cs.storage.SavePrivateFile(v, ext, "chat_attachment")
		if err != nil {
			cs.cleanUpAttachment(chatAttachments)
			return nil, &customerrors.ServiceErrors{
				Code:    http.StatusInternalServerError,
				Message: "Gagal saat menyimpan file  " + err.Error(),
			}
		}

		attObj := model.ChatAttachment{
			FileName:  fName,
			MediaType: cs.storage.GetMediaType(mimeType),
			Size:      v.Size,
			ChatId:    chatId,
		}

		chatAttachments = append(chatAttachments, attObj)

	}

	return chatAttachments, nil

}

func (cs *ChatService) sendChat(d *ChatPostInput, att []model.ChatAttachment) {
	receiverRoom := "user:" + d.ReceiverId
	media := []string{}

	for _, v := range att {
		media = append(media, v.FileName)
	}

	payloadData, _ := json.Marshal(ws.PrivateMessageData{
		Message:  d.ChatText,
		MediaUrl: media,
	})

	// todo ini testing doang
	payload, _ := json.Marshal(ws.WebsocketEvent{
		Action: ws.ActionPrivateMessage,
		Detail: "NEW MESSSAGE ARRIVED",
		Type:   ws.TypeSystemOk,
		Data:   payloadData,
	})

	cs.Hub.SendPayloadTo(receiverRoom, payload)

}

func parseToChatModel(cp *ChatPostInput) (model.ChatModel, error) {

	receiverUuid, err := uuid.Parse(cp.ReceiverId)

	if err != nil {
		return model.ChatModel{}, err
	}

	replyTo, err := pkg.StringToUuid(cp.ReplyTo)
	if err != nil {
		return model.ChatModel{}, err
	}

	postId, err := pkg.StringToUuid(cp.PostId)
	if err != nil {
		return model.ChatModel{}, err
	}

	return model.ChatModel{
		SenderId:   cp.SenderId,
		ReceiverId: receiverUuid,
		ReplyTo:    replyTo,
		ChatText:   cp.ChatText,
		PostId:     postId,
		IsRead:     false,
	}, nil

}

func (cs *ChatService) cleanUpAttachment(list []model.ChatAttachment) {

	if len(list) > 0 {
		for _, v := range list {
			cs.storage.DeletePrivateFile(v.FileName, "chat_attachment")
		}
	}

}
