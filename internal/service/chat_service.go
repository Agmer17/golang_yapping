package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

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
}

func NewChatService(c *repository.ChatRepository,
	h *ws.Hub,
	u *UserService,
	ct *repository.ChatAttachmentRepository) *ChatService {
	return &ChatService{
		Pool:    c,
		Hub:     h,
		usv:     u,
		chatAtt: ct,
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

	savedChat, err := cs.Pool.Save(cm)

	if len(d.MediaFiles) != 0 {
		// todo : olah file nya
		listMetadata, svcErr := processAttachment(d.MediaFiles, savedChat.Id)
		if svcErr != nil {

			return svcErr
		}
		fmt.Println(listMetadata)

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

func processAttachment(att []*multipart.FileHeader, chatId uuid.UUID) ([]model.ChatAttachment, *customerrors.ServiceErrors) {

	var chatAttachments []model.ChatAttachment = make([]model.ChatAttachment, 0)

	for _, v := range att {
		mimeType, err := pkg.DetectFileType(v)
		if err != nil {
			return nil, &customerrors.ServiceErrors{
				Code:    http.StatusInternalServerError,
				Message: "Terjadi kesalahan saat menyimpan file " + err.Error(),
			}
		}

		ext, ok := pkg.IsTypeSupportted(mimeType)

		if !ok {
			return nil, &customerrors.ServiceErrors{
				Code:    http.StatusBadRequest,
				Message: "File saat ini tidak didukung!",
			}

		}

		fName, err := pkg.SavePrivateFile(v, ext)

		if err != nil {
			return nil, &customerrors.ServiceErrors{
				Code:    http.StatusInternalServerError,
				Message: "Gagal saat menyimpan file  " + err.Error(),
			}
		}

		attObj := model.ChatAttachment{
			FileName:  fName,
			MediaType: getMediaType(mimeType),
			Size:      v.Size,
			ChatId:    chatId,
		}

		chatAttachments = append(chatAttachments, attObj)

	}

	return chatAttachments, nil

}

func getMediaType(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "IMAGE"
	case strings.HasPrefix(mime, "video/"):
		return "VIDEO"
	case strings.HasPrefix(mime, "audio/"):
		return "AUDIO"
	case strings.HasPrefix(mime, "application/"):
		return "DOCUMENT"
	default:
		return ""
	}
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
