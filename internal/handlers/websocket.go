package handlers

import (
	"fmt"
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/Agmer17/golang_yapping/pkg"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketHandler struct {
	Hub      *ws.Hub
	Upgrader websocket.Upgrader
}

func NewWebsocketHandler() *WebSocketHandler {

	return &WebSocketHandler{
		Hub: ws.NewHub(),
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

}
func (w *WebSocketHandler) RegisterRoutes(r *gin.RouterGroup) {

	we := r.Group("/ws")

	{
		we.GET("/connect", w.ServeWS)
	}

}

func (w *WebSocketHandler) ServeWS(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	token, err := pkg.GetAccessToken(authHeader)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Harap login terlebih dahulu sebelum mengakses fitur ini!",
		})
		return
	}

	fmt.Println("\n\n\n\n\n\n token : ", token)

	claims, err := pkg.VerifyToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token kadaluarsa atau tidak valid"})
		c.Abort()
		return
	}
	fmt.Println("\n\n\n\n\n\n claims : ", claims)
	userId := claims.UserID

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	room := w.Hub.GetOrCreate(userId.String())
	client := ws.NewClient(conn, room, userId)

	room.Register <- client

	go client.WritePump()
	client.ReadPump()
}
