package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
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

	c.JSON(http.StatusOK, gin.H{"data": rBind, "msg": "berhasil!"})

}
