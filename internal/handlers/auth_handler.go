package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignUpRequest struct {
	Username string `json:"username" binding:"required,min=4,max=150"`
	Email    string `json:"email" binding:"required,email"`
	Fullname string `json:"full_name" binding:"required,min=4,max=150"`
	Password string `json:"password" binding:"required,min=5,max=150"`
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
		auth.POST("/sign-up", h.handleSignUp)
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

	sevenDays := time.Hour * 24 * 7

	c.SetCookie("refreshToken",
		resp["refreshToken"].(string),
		int(sevenDays), "/",
		"",
		true,
		true)

	c.JSON(http.StatusAccepted, gin.H{"message": "berhasil",
		"accessToken": resp["accessToken"],
		"id":          resp["id"],
	})
}

func (h *AuthHandler) handleSignUp(c *gin.Context) {

	var sBind SignUpRequest

	err := c.ShouldBindJSON(&sBind)

	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			errorsMap := make(map[string]string)
			for _, e := range validationErrors {
				errorsMap[e.Field()] = fmt.Sprintf("failed on '%s' tag", e.Tag())
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": errorsMap})
			return
		}

	}

	resp, customErr := h.Service.SignUp(sBind.Username, sBind.Email, sBind.Fullname, sBind.Password, c.Request.Context())
	if customErr != nil {
		c.JSON(customErr.Code, gin.H{"message": customErr.Message})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
