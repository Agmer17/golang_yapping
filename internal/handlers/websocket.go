package handlers

import (
	"net/http"

	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketHandler struct {
	Hub      *ws.Hub
	Upgrader websocket.Upgrader
}

func NewWebsocketHandler(h *ws.Hub) *WebSocketHandler {

	return &WebSocketHandler{
		Hub: h,
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

	val, ok := c.Get("userId")

	if !ok {
		c.JSON(401, gin.H{
			"error": "harap login sebelum mengakses ini!",
		})
		return
	}

	userId := val.(uuid.UUID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	room := w.Hub.GetOrCreate("user:" + userId.String())
	client := ws.NewClient(conn, room, userId)

	room.Register <- client

	go client.WritePump()
	client.ReadPump()
}
