package middleware

import (
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		header := ctx.GetHeader("Authorization")
		accessToken, err := pkg.GetAccessToken(header)

		if err != nil {
			ctx.JSON(401, gin.H{
				"error": err.Error(),
			})
			ctx.Abort()
			return
		}

		accesClaims, err := pkg.VerifyToken(accessToken)
		if err != nil {
			ctx.JSON(401, gin.H{
				"error": err.Error(),
			})
			ctx.Abort()
			return
		}

		if accesClaims != nil {
			ctx.Set("userId", accesClaims.UserID)
			ctx.Next()
			return

		} else {
			ctx.JSON(401, gin.H{
				"error": "harap login sebelum mengakses fitur ini",
			})
			ctx.Abort()
			return

		}
	}

}
