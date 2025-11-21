package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebsocketHandler struct {
	Upgrader websocket.Upgrader
}

func NewWebsocketHandler() *WebsocketHandler {

	return &WebsocketHandler{
		Upgrader: websocket.Upgrader{},
	}

}

func (w *WebsocketHandler) RegisterRoutes(rg *gin.RouterGroup) {
	ws := rg.Group("/ws")

	{
		ws.GET("/", w.serveWebsocket)
	}

}

func (w WebsocketHandler) serveWebsocket(c *gin.Context) {

	writer := c.Writer
	request := c.Request

	conn, err := w.Upgrader.Upgrade(writer, request, nil)

	if err != nil {
		log.Print("error saat menghubungkan dengan websocket", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "gagal mengupgrade koneksi ke websocket" + err.Error()})

	}

	fmt.Println("berhasil terhubung ke websocket")

	defer conn.Close()
}
