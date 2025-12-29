package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/repository"
	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/Agmer17/golang_yapping/pkg/customerrors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// internal struct!
type ChatPostInput struct {
	SenderId   uuid.UUID
	ReceiverId string
	ReplyTo    *string
	ChatText   *string
	PostId     *string
	MediaFiles []*multipart.FileHeader
}

// key buat get ini di redis tuh media_access:private_chat:<token>
type mediaAccessToken struct {
	Filename   string    `redis:"filename"`
	SenderId   uuid.UUID `redis:"sender_id"`
	ReceiverId uuid.UUID `redis:"receiver_id"`
}

type ChatResponseData struct {
	Id               uuid.UUID  `json:"chat_id"`
	SenderId         uuid.UUID  `json:"sender_id"`
	ReceiverId       uuid.UUID  `json:"receiver_id"`
	ReplyTo          *uuid.UUID `json:"reply_to"`
	ChatText         *string    `json:"chat_text"`
	PostId           *uuid.UUID `json:"post_id"`
	IsRead           bool       `json:"is_read"`
	CreatedAt        time.Time  `json:"created_at"`
	IsOwn            bool       `json:"is_own_message"`
	AttachmentAccess []string   `json:"attachment_access"`
}

// =======

type ChatServiceInterface interface {
	SaveChat(m *ChatPostInput, ctx context.Context) *customerrors.ServiceErrors
	GetChatBeetween(ctx context.Context, r uuid.UUID, s uuid.UUID) ([]ChatResponseData, *customerrors.ServiceErrors)
	GetPrivateAttachmentFile(ctx context.Context, key string, userId uuid.UUID) (string, *customerrors.ServiceErrors)
}

type ChatService struct {
	Pool        repository.ChatRepositoryInterface
	Hub         *ws.Hub
	usv         *UserService
	chatAtt     repository.ChatAttachmentInterface
	storage     *FileStorage
	RedisClient *redis.Client
}

func NewChatService(c *repository.ChatRepository,
	h *ws.Hub,
	u *UserService,
	ct *repository.ChatAttachmentRepository,
	fileService *FileStorage,
	redisCli *redis.Client) *ChatService {
	return &ChatService{
		Pool:        c,
		Hub:         h,
		usv:         u,
		chatAtt:     ct,
		storage:     fileService,
		RedisClient: redisCli,
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

		savedChat.Attachment = listMetadata

		tokenAcc, err := cs.setTokenToAccess(ctx, savedChat.Attachment, savedChat.SenderId, savedChat.ReceiverId)

		if err != nil {
			return &customerrors.ServiceErrors{
				Code:    500,
				Message: "Terjadi kesalahan di server! " + err.Error(),
			}
		}

		go cs.sendChat(savedChat, tokenAcc)
		return nil
	}

	go cs.sendChat(savedChat, []string{})
	return nil
}

func (cs *ChatService) GetChatBeetween(ctx context.Context, receiver uuid.UUID, sender uuid.UUID) ([]ChatResponseData, *customerrors.ServiceErrors) {

	chatList, err := cs.Pool.GetChatBeetween(ctx, receiver, sender)

	if err != nil {
		log.Print("terjadi error saat mengambil data di datbase : " + err.Error() + "\n")
		return nil, &customerrors.ServiceErrors{
			Code:    500,
			Message: "error " + err.Error(),
		}
	}

	// todo buat ini di redis!
	data, err := cs.setToChatResponse(ctx, &chatList, sender)

	return data, nil

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

func (cs *ChatService) sendChat(savedChat model.ChatModel, attData []string) {

	destRoom := "user:" + savedChat.ReceiverId.String()

	pvData := ws.PrivateMessageData{
		Message:  savedChat.ChatText,
		MediaUrl: attData,
	}

	pvDataByte, _ := json.Marshal(pvData)

	wsEvent := ws.WebsocketEvent{
		Action: ws.ActionPrivateMessage,
		Detail: "NEW MESSAGE ARRIVED",
		Type:   ws.TypeSystemOk,
		Data:   pvDataByte,
	}

	wsEventByte, _ := json.Marshal(wsEvent)

	cs.Hub.SendPayloadTo(destRoom, wsEventByte)

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

func (cs *ChatService) getAttachmentFromToken(ctx context.Context, key string) (mediaAccessToken, error) {

	var token mediaAccessToken

	cmd := cs.RedisClient.HGetAll(ctx, key)
	if err := cmd.Err(); err != nil {
		return mediaAccessToken{}, err
	}

	data := cmd.Val()
	// fmt.Println("\n\n\n\n\n data val : " + cmd.Val()["filename"] + "tokennya : ")
	if len(data) == 0 {
		return mediaAccessToken{}, &customerrors.ServiceErrors{Code: 404, Message: "token tidak valid atau tidak ditemukan"}
	}

	scanErr := cmd.Scan(&token)

	if scanErr != nil {
		return mediaAccessToken{}, scanErr
	}

	// fmt.Println("\n\n\n\n\n data token : " + token.UserId.String() + "tokennya : ")

	return token, nil
}

func (cs *ChatService) setTokenToAccess(ctx context.Context, att []model.ChatAttachment, sender uuid.UUID, receiverId uuid.UUID) ([]string, error) {

	var listTempData []map[string]any

	var tokenList []string

	for _, v := range att {
		listTempData = append(listTempData, map[string]any{
			"filename":    v.FileName,
			"sender_id":   sender.String(),
			"receiver_id": receiverId.String(),
		})

	}

	pipe := cs.RedisClient.Pipeline()

	for _, data := range listTempData {
		token, err := pkg.GenerateRandomStringToken(16)
		if err != nil {
			return nil, err
		}

		key := "media_access:private_chat:" + token

		pipe.HSet(ctx, key, data)

		pipe.Expire(ctx, key, 45*time.Second)

		tokenList = append(tokenList, token)
	}

	_, err := pipe.Exec(ctx)

	if err != nil {
		return nil, err
	}

	return tokenList, nil
}

func (cs *ChatService) setToChatResponse(ctx context.Context, data *[]model.ChatModel, userId uuid.UUID) ([]ChatResponseData, error) {

	var ResponseList []ChatResponseData

	for _, val := range *data {

		tmpResp := ChatResponseData{
			Id:         val.Id,
			SenderId:   val.SenderId,
			ReceiverId: val.ReceiverId,
			ReplyTo:    val.ReplyTo,
			ChatText:   val.ChatText,
			PostId:     val.PostId,
			IsRead:     val.IsRead,
			CreatedAt:  val.CreatedAt,
			IsOwn:      val.IsOwn,
		}

		if len(val.Attachment) == 0 {
			tmpResp.AttachmentAccess = []string{}
		} else {
			tmpToken, err := cs.setTokenToAccess(ctx, val.Attachment, val.SenderId, val.ReceiverId)

			if err != nil {
				return nil, err
			}

			tmpResp.AttachmentAccess = tmpToken
		}

		ResponseList = append(ResponseList, tmpResp)
	}

	return ResponseList, nil

}

func (cs *ChatService) GetPrivateAttachmentFile(ctx context.Context, key string, userId uuid.UUID) (string, *customerrors.ServiceErrors) {

	var svcError *customerrors.ServiceErrors
	mediaAccess, err := cs.getAttachmentFromToken(ctx, key)

	if err != nil {
		if errors.As(err, &svcError) {
			return "", svcError
		} else {
			return "", &customerrors.ServiceErrors{
				Code:    500,
				Message: "internal server error! " + err.Error(),
			}
		}

	}

	if mediaAccess.SenderId != userId && mediaAccess.ReceiverId != userId {
		// fmt.Println("salah user id! harusnya : " + mediaAccess.SenderId.String() + "atau : " + userId.String())
		return "", &customerrors.ServiceErrors{
			Code:    401,
			Message: "Unauthorized access! kamu tidak berhak mengkases file ini!",
		}
	}

	return cs.storage.GetPathPrivateFile(mediaAccess.Filename, "chat_attachment"), nil

}
