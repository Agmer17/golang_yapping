package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/Agmer17/golang_yapping/pkg"
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
		auth.GET("/refresh-session", h.refreshSession)
		auth.GET("/activate-account/:token", h.handleActivateAccount)
	}

}

func (h *AuthHandler) handleLogin(c *gin.Context) {

	var rBind loginRequest

	err := c.ShouldBindJSON(&rBind)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request! isi data degan benar"})
		c.Abort()
		return
	}

	resp, serviceErr := h.Service.LoginService(rBind.Username, rBind.Password, c.Request.Context())
	if serviceErr != nil {
		c.JSON(serviceErr.Code, gin.H{"error": serviceErr.Message})
		c.Abort()
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
			c.Abort()
			return
		}

	}

	resp, customErr := h.Service.SignUp(sBind.Username, sBind.Email, sBind.Fullname, sBind.Password, c.Request.Host, c.Request.Context())
	if customErr != nil {
		c.JSON(customErr.Code, gin.H{"error": customErr.Message})
		c.Abort()
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) refreshSession(c *gin.Context) {

	refreshToken, err := c.Cookie("refreshToken")

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Harap login terlebih dahulu sebelum mengakses fitur ini!",
		})
		c.Abort()
		return
	}

	accessToken, serviceErr := h.Service.RefreshSession(refreshToken, c.Request.Context())

	if serviceErr != nil {
		c.JSON(serviceErr.Code, gin.H{
			"error": serviceErr.Message,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, accessToken)
}

func (h *AuthHandler) handleActivateAccount(c *gin.Context) {

	token := c.Param("token")

	if pkg.IsPStrEmpty(&token) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parameter token tidak valid, harap masukan url dengan benar",
		})
		return

	}

	// todo : impl ngecek tokennya mas

}
