package handlers

import (
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthHandler struct {
	Service service.AuthServiceInterface
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		Service: svc,
	}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {

	auth := rg.Group("/auth")

	{
		auth.POST("/login", h.handleLogin)
	}

}

func (h *AuthHandler) handleLogin(c *gin.Context) {

	var rBind loginRequest

	err := c.ShouldBindJSON(&rBind)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request! isi data degan benar"})
		return
	}

	resp, serviceErr := h.Service.LoginService(rBind.Username, rBind.Password, c.Request.Context())
	if serviceErr != nil {
		c.JSON(serviceErr.Code, gin.H{"error": serviceErr.Message})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "berhasil", "data": resp})
}
