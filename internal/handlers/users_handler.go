package handlers

import (
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc service.UserServiceInterface
}

func (u *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {

	user := rg.Group("/user")

	{
		user.GET("/me", u.handleMyProfile)
	}

}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{
		svc: s,
	}
}

func (u *UserHandler) handleMyProfile(c *gin.Context) {

	val, ok := c.Get("userId")
	if !ok {
		c.JSON(401, gin.H{
			"error": "harap login sebelum mengakses ini!",
		})
		return
	}

	userId := val.(uuid.UUID)

	data, err := u.svc.GetMyProfile(userId, c.Request.Context())

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Data tidak ditemukan atau terjadi kesalahan di server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "berhasil mengambil data",
	})

}
