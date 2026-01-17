package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/Agmer17/golang_yapping/internal/event"
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

type LatestChatData struct {
	ChatResponseData      `json:"chat_data"`
	ChatPartnerId         uuid.UUID `json:"partner_id"`
	ChatPartnerFullName   string    `json:"partner_fullname"`
	ChatPartnerUsername   string    `json:"partner_username"`
	ChatPartnerProfilePic *string   `json:"partner_profile_picture"`
}

// =======

type ChatServiceInterface interface {
	SaveChat(m *ChatPostInput, ctx context.Context) *customerrors.ServiceErrors
	GetChatBeetween(ctx context.Context, r uuid.UUID, s uuid.UUID) ([]ChatResponseData, *customerrors.ServiceErrors)
	GetPrivateAttachmentFile(ctx context.Context, key string, userId uuid.UUID) (string, *customerrors.ServiceErrors)
	GetLatestChat(ctx context.Context, userId uuid.UUID) ([]LatestChatData, *customerrors.ServiceErrors)
	DeleteChat(ctx context.Context, userId uuid.UUID, chatId uuid.UUID) *customerrors.ServiceErrors
}

type ChatService struct {
	Pool        repository.ChatRepositoryInterface
	Hub         *ws.Hub
	usv         *UserService
	chatAtt     repository.ChatAttachmentInterface
	storage     *FileStorage
	RedisClient *redis.Client
	EventBus    *event.EventBus
}

func NewChatService(c *repository.ChatRepository,
	h *ws.Hub,
	u *UserService,
	ct *repository.ChatAttachmentRepository,
	fileService *FileStorage,
	redisCli *redis.Client,
	eventBus *event.EventBus) *ChatService {
	return &ChatService{
		Pool:        c,
		Hub:         h,
		usv:         u,
		chatAtt:     ct,
		storage:     fileService,
		RedisClient: redisCli,
		EventBus:    eventBus,
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
		listMetadata, svcErr := cs.processAttachment(d.MediaFiles, savedChat.ChatData.Id)
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

		savedChat.ChatData.Attachment = listMetadata

		tokenAcc, err := cs.setTokenToAccess(ctx, savedChat.ChatData.Attachment, savedChat.ChatData.SenderId, savedChat.ChatData.ReceiverId)

		if err != nil {
			return &customerrors.ServiceErrors{
				Code:    500,
				Message: "Terjadi kesalahan di server! " + err.Error(),
			}
		}

		cs.sendChat(savedChat, tokenAcc)
		return nil
	}

	cs.sendChat(savedChat, []string{})
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
	data, err := cs.setToChatResponses(ctx, chatList, sender)

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

func (cs ChatService) sendChat(savedChat repository.ChatWithSender, attData []string) {

	destRoom := "user:" + savedChat.ChatData.ReceiverId.String()
	senderDestRoom := "user:" + savedChat.ChatData.SenderId.String()

	pvData := ws.PrivateMessageData{
		Message:  savedChat.ChatData.ChatText,
		MediaUrl: attData,
	}

	pvData.From = ws.UserMetadata{
		Id:             savedChat.Sender.Id,
		Username:       savedChat.Sender.Username,
		FullName:       savedChat.Sender.FullName,
		ProfilePicture: savedChat.Sender.ProfilePicture,
	}

	pvDataByte, _ := json.Marshal(pvData)

	wsEvent := ws.WebsocketEvent{
		Action:   ws.ActionPrivateMessage,
		Detail:   "NEW MESSAGE ARRIVED",
		Type:     ws.TypeSystemOk,
		Receiver: destRoom,
		Data:     pvDataByte,
	}

	senderWsEvent := ws.WebsocketEvent{
		Action:   ws.ActionPrivateMessage,
		Detail:   "MESSAGE SUCCESSFULLY DELIVERED",
		Type:     ws.TypeSystemOk,
		Receiver: senderDestRoom,
		Data:     pvDataByte,
	}

	cs.EventBus.Publish(event.WsEventSendPayload, wsEvent)

	cs.EventBus.Publish(event.WsEventSendPayload, senderWsEvent)
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

	return token, nil
}

func (cs *ChatService) setTokenToAccess(
	ctx context.Context,
	att []model.ChatAttachment,
	sender uuid.UUID,
	receiverId uuid.UUID,
) ([]string, error) {

	pipe := cs.RedisClient.Pipeline()
	tokenList := make([]string, 0, len(att))

	for _, v := range att {
		token, err := pkg.GenerateRandomStringToken(16)
		if err != nil {
			return nil, err
		}

		key := "media_access:private_chat:" + token

		data := map[string]string{
			"filename":    v.FileName,
			"sender_id":   sender.String(),
			"receiver_id": receiverId.String(),
			"type":        v.MediaType,
		}

		pipe.HSet(ctx, key, data)

		ttl := cs.resolveMediaTTL(v.MediaType, v.FileName)
		pipe.Expire(ctx, key, ttl)

		tokenList = append(tokenList, token)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return tokenList, nil
}

func (cs *ChatService) setToChatResponses(ctx context.Context, data []model.ChatModel, userId uuid.UUID) ([]ChatResponseData, error) {

	var ResponseList []ChatResponseData

	for _, val := range data {

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

func (cs *ChatService) setOneToChatResponse(ctx context.Context, data model.ChatModel, userId uuid.UUID) (ChatResponseData, error) {

	tmpResp := ChatResponseData{
		Id:         data.Id,
		SenderId:   data.SenderId,
		ReceiverId: data.ReceiverId,
		ReplyTo:    data.ReplyTo,
		ChatText:   data.ChatText,
		PostId:     data.PostId,
		IsRead:     data.IsRead,
		CreatedAt:  data.CreatedAt,
		IsOwn:      data.IsOwn,
	}

	if len(data.Attachment) == 0 {
		tmpResp.AttachmentAccess = []string{}
	} else {
		tmpToken, err := cs.setTokenToAccess(ctx, data.Attachment, data.SenderId, data.ReceiverId)

		if err != nil {
			return ChatResponseData{}, err
		}

		tmpResp.AttachmentAccess = tmpToken
	}

	return tmpResp, nil

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
		return "", &customerrors.ServiceErrors{
			Code:    401,
			Message: "Unauthorized access! kamu tidak berhak mengkases file ini!",
		}
	}

	return cs.storage.GetPathPrivateFile(mediaAccess.Filename, "chat_attachment"), nil

}

func (cs *ChatService) GetLatestChat(ctx context.Context, userId uuid.UUID) ([]LatestChatData, *customerrors.ServiceErrors) {

	data, err := cs.Pool.GetLastChat(ctx, userId)
	if err != nil {
		return []LatestChatData{}, &customerrors.ServiceErrors{
			Code:    500,
			Message: err.Error(),
		}
	}

	// fmt.Println(data)
	var LatestChats []LatestChatData = make([]LatestChatData, 0)
	for _, v := range data {
		tmpChatResp, err := cs.setOneToChatResponse(ctx, v.ChatData, userId)

		if err != nil {
			return nil, &customerrors.ServiceErrors{
				Code:    500,
				Message: "Gagal saat mengambil data dari database " + err.Error(),
			}
		}

		tmpRespData := LatestChatData{
			ChatResponseData:      tmpChatResp,
			ChatPartnerId:         v.PartnerData.Id,
			ChatPartnerFullName:   v.PartnerData.FullName,
			ChatPartnerUsername:   v.PartnerData.Username,
			ChatPartnerProfilePic: v.PartnerData.ProfilePicture,
		}

		LatestChats = append(LatestChats, tmpRespData)

	}

	return LatestChats, nil
}

func (cs *ChatService) DeleteChat(ctx context.Context, userId uuid.UUID, chatId uuid.UUID) *customerrors.ServiceErrors {

	chatMetaData, err := cs.Pool.GetChatById(ctx, chatId)

	if err != nil {

		if err == sql.ErrNoRows {

			return &customerrors.ServiceErrors{
				Code:    404,
				Message: "Chat tidak ditemukan!",
			}
		} else {
			return &customerrors.ServiceErrors{
				Code:    500,
				Message: "Gagal mengahpus data di database " + err.Error(),
			}
		}

	}

	if chatMetaData.SenderId != userId {
		return &customerrors.ServiceErrors{
			Code:    401,
			Message: "Kamu tidak bisa menghapus pesan ini!",
		}
	}

	filesToDelete, err := cs.Pool.Delete(ctx, chatId)

	if err != nil {
		return &customerrors.ServiceErrors{
			Code:    500,
			Message: "gagal menghapus chat " + err.Error(),
		}
	}

	cs.storage.DeleteAllPrivateFile(filesToDelete, "chat_attachment")
	return nil

}

func (cs *ChatService) resolveMediaTTL(
	mediaType string,
	filename string,
) time.Duration {

	const defaultTTL = 5 * time.Minute

	if mediaType != model.TypeVideo {
		return defaultTTL
	}

	duration, err := cs.storage.GetVideoDurationPVT(filename, "chat_attachment")
	if err != nil {
		fmt.Println("get video duration error:", err)
		return defaultTTL
	}

	// safeguard minimum TTL
	if duration < defaultTTL {
		return defaultTTL
	}

	return duration
}
