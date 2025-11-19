package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc service.UserServiceInterface
}

func (u *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {

	user := rg.Group("/user")

	user.Use()

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

	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Harap login sebelum mengakses ini",
		})
		c.Abort()
		return

	}
	parts := strings.SplitN(authHeader, " ", 2)

	fmt.Println("parts ", parts)

	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Header otorisasi tidak valid!"})
		c.Abort()
		return
	}

	claims, err := pkg.VerifyToken(parts[1])

	fmt.Println("claims : ", claims)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token kadaluarsa atau tidak valid"})
		c.Abort()
		return
	}

	data, err := u.svc.GetMyProfile(claims.UserID, c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "berhasil mengambil data",
	})

}
