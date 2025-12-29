package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

const MaxFileSizes = 5 << 20

type ChatHandler struct {
	svc service.ChatServiceInterface
}

type PostChatRequest struct {
	ReceiverId string                  `form:"receiver_id" binding:"required,uuid"`
	ReplyTo    *string                 `form:"reply_to" binding:"omitempty,uuid"`
	ChatText   *string                 `form:"chat_text"`
	PostId     *string                 `form:"posts_id" binding:"omitempty,uuid"`
	MediaFiles []*multipart.FileHeader `form:"chat_media"`
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{
		svc: svc,
	}
}

func (chat *ChatHandler) RegisterRoutes(rg *gin.RouterGroup) {

	chatEndpoint := rg.Group("/chat")

	{
		chatEndpoint.POST("/post-message", chat.PostChat)
		chatEndpoint.GET("/beetween/:receiver", chat.GetChatBeetween)
		chatEndpoint.GET("/attachment/:token", chat.GetChatAttachment)
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

	for _, file := range pc.MediaFiles {
		if file.Size > MaxFileSizes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File %s terlalu besar, maksimal 5MB", file.Filename),
			})
			return
		}
	}

	postInput := service.ChatPostInput{
		SenderId:   currentUser,
		ReceiverId: pc.ReceiverId,
		ReplyTo:    pc.ReplyTo,
		ChatText:   pc.ChatText,
		MediaFiles: pc.MediaFiles,
		PostId:     pc.PostId,
	}

	svcErr := chat.svc.SaveChat(&postInput, c.Request.Context())
	if svcErr != nil {
		c.JSON(svcErr.Code, gin.H{
			"error": svcErr.Message,
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func (chat *ChatHandler) GetChatBeetween(c *gin.Context) {

	val, ok := c.Get("userId")
	if !ok {
		c.JSON(401, gin.H{
			"error": "harap login sebelum mengakses ini!",
		})
		return
	}

	receiverString := c.Param("receiver")

	senderId := val.(uuid.UUID)

	receiverId, err := uuid.Parse(receiverString)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parameter tidak valid! harap masukan parameter dengan benar",
		})
	}

	data, svcErr := chat.svc.GetChatBeetween(c.Request.Context(), receiverId, senderId)
	if svcErr != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "chat atau user tidak ditemukan! " + svcErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "berhasil mengambil data!",
		"data":    data,
	})

}

func (chat *ChatHandler) GetChatAttachment(c *gin.Context) {

}
