package handlers

import (
	"fmt"
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

type ChatHandler struct {
	svc service.ChatServiceInterface
}

type PostChatRequest struct {
	ReceiverId string  `form:"receiver_id" binding:"required,uuid"`
	ReplyTo    *string `form:"reply_to" binding:"omitempty,uuid"`
	ChatText   *string `form:"chat_text"`
	PostId     *string `form:"posts_id" binding:"omitempty,uuid"`
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{
		svc: svc,
	}
}

func (chat *ChatHandler) RegisterRoutes(rg *gin.RouterGroup) {

	chatEndpoint := rg.Group("/chat")

	{
		chatEndpoint.POST("/private-message", chat.PostChat)
	}

}

func (chat *ChatHandler) PostChat(c *gin.Context) {
	val, ok := c.Get("userId")
	if !ok {
		c.JSON(401, gin.H{
			"error": "harap login sebelum mengakses ini!",
		})
		return
	}
	currentUser := val.(uuid.UUID)

	var pc PostChatRequest

	if err := c.ShouldBindWith(&pc, binding.FormMultipart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Harap isi data dengan benar!",
		})
		return
	}

	file, err := c.FormFile("chat_media")

	receiverId, _ := uuid.Parse(pc.ReceiverId)
	var repTo uuid.UUID
	var postId uuid.UUID

	if pc.ReplyTo != nil && *pc.ReplyTo != "" {
		repTo, _ = uuid.Parse(*pc.ReplyTo)
	}

	if pc.PostId != nil && *pc.PostId != "" {
		postId, _ = uuid.Parse(*pc.PostId)
	}

	// ini maksudnya kalo file kosong/gak ada media
	if err != nil {

		cm := model.ChatModel{
			SenderId:   currentUser,
			ReceiverId: receiverId,
			ReplyTo:    repTo,
			PostId:     postId,

			ChatText: pc.ChatText,
		}

		fmt.Println(cm)

		svErr := chat.svc.SaveChat(&cm)

		if svErr != nil {
			c.JSON(svErr.Code, gin.H{
				"error": svErr.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "berhasil",
		})

		return

	} else {
		// ini kalo file nya ada
		mimeType, err := pkg.DetectFileType(file)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error saat parsing file : " + err.Error(),
			})
			return
		}

		fileType, ok := pkg.IsValidImage(mimeType)

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "format file ini di chat belum di dukung",
			})
			return
		}

		fileName, err := pkg.SavePrivateFile(file, fileType)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "gagal menyimpan file : " + err.Error(),
			})
			return
		}

		cm := model.ChatModel{
			SenderId:   currentUser,
			ReceiverId: receiverId,
			ReplyTo:    repTo,
			PostId:     postId,

			ChatText:  pc.ChatText,
			ChatMedia: &fileName,
		}

		svErr := chat.svc.SaveChat(&cm)

		if svErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": svErr.Error(),
			})
			return

		}

		c.JSON(200, gin.H{
			"message": "berhasil mengirim pesan",
		})

	}

}
