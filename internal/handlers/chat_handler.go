package handlers

import (
	"fmt"
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	svc service.ChatServiceInterface
}

type PostChatRequest struct {
	ReceiverId uuid.UUID  `form:"receiver_id" binding:"required"`
	ReplyTo    *uuid.UUID `form:"reply_to"`
	ChatText   string     `form:"chat_text"`
	PostId     *uuid.UUID `form:"posts_id"`
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

	fmt.Println(currentUser)

	receiverID, err := uuid.Parse(c.PostForm("receiver_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "receiver_id tidak valid!"})
		return
	}

	var replyTo *uuid.UUID
	if rt := c.PostForm("reply_to"); rt != "" {
		parsed, _ := uuid.Parse(rt)
		replyTo = &parsed
	}

	var postID *uuid.UUID
	if pid := c.PostForm("posts_id"); pid != "" {
		parsed, _ := uuid.Parse(pid)
		postID = &parsed
	}

	pc := PostChatRequest{
		ReceiverId: receiverID,
		ReplyTo:    replyTo,
		PostId:     postID,
		ChatText:   c.PostForm("chat_text"),
	}

	fmt.Println(pc)

	file, err := c.FormFile("chat_media")

	if err != nil {

		c.JSON(http.StatusOK, gin.H{
			"message": "ok : ",
		})

		return

	}

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
			"error": "saat ini format file di chat belum di dukung",
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

	c.JSON(200, gin.H{
		"message": "ok : " + fileName,
	})

}
