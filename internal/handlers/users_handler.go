package handlers

import (
	"fmt"
	"net/http"

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

	token, err := pkg.GetAccessToken(authHeader)

	if err != nil {
		fmt.Println("\n\n\n\n\n err : ", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token kadaluarsa atau tidak valid"})
		c.Abort()
		return
	}

	fmt.Println("\n\n\n\n token : ", token)
	claims, err := pkg.VerifyToken(token)

	fmt.Println("claims : ", claims)

	if err != nil {
		fmt.Println("\n\n\n\n\n err : ", err.Error())

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
