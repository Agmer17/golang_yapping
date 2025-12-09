// client.go
package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 1024
)

type Client struct {
	Conn   *websocket.Conn
	Send   chan []byte
	Room   *Room
	UserId uuid.UUID
}

func NewClient(conn *websocket.Conn, room *Room, userId uuid.UUID) *Client {

	return &Client{
		Conn:   conn,
		Send:   make(chan []byte),
		Room:   room,
		UserId: userId,
	}

}

func (c *Client) ReadPump() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var jsonEvent WebsocketEvent

		err = json.Unmarshal(message, &jsonEvent)

		if err != nil {
			c.sendError("Payload tidak didukung!")
			continue
		}

		switch jsonEvent.Action {

		case actionSubscribe:
			var joinData JoinRoomEventData
			if err := json.Unmarshal(jsonEvent.Data, &joinData); err != nil {
				c.sendError("Harap kirim payload dengan benar!")
				continue
			}

			if err := binding.Validator.ValidateStruct(joinData); err != nil {

				c.sendError("Harap kirim data payload dengan benar")
				continue
			}

			c.Room.Broadcast <- message

		case actionPrivateMessage:

			c.processPrivateMessage(jsonEvent)

		default:
			c.sendError("event not supported")
		}
	}
}

func (c *Client) WritePump() {

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {

		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				errPayload, _ := json.Marshal(WebsocketEvent{
					Action: "SYSTEM",
					Detail: "KONEKSI DITUTUP KARENA INTERNET MELAMBAT!",
					Data:   nil,
				})

				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				c.Conn.WriteMessage(websocket.CloseMessage, errPayload)
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}

	}

}

func (c *Client) sendError(msg string) {

	errEvent, _ := json.Marshal(WebsocketEvent{
		Action: actionSystem,
		Detail: msg,
		Type:   typeSystemError,
		Data:   nil,
	})

	c.Send <- errEvent

}

func (c *Client) processPrivateMessage(jsonEvent WebsocketEvent) {

	var pvmsg PrivateMessageData

	if err := json.Unmarshal(jsonEvent.Data, &pvmsg); err != nil {
		c.sendError("Terjadi kesalahan saat parsing payload")
		return
	}

	if err := binding.Validator.ValidateStruct(pvmsg); err != nil {
		c.sendError("Harap kirim payload dengan benar!")
		return
	}

	receiver := c.Room.Hub.GetRoom("user:" + pvmsg.To.String())

	resData, err := json.Marshal(PrivateMessageData{
		To:        pvmsg.To,
		Message:   pvmsg.Message,
		Media_url: pvmsg.Media_url,
		From:      c.UserId,
	})

	if err != nil {
		c.sendError("Gagal memproses pesan")
		return
	}

	response, err := json.Marshal(WebsocketEvent{
		Action: jsonEvent.Action,
		Detail: jsonEvent.Detail,
		Type:   typeSystemOk,
		Data:   resData,
	})

	if err != nil {
		c.sendError("Gagal memproses pesan!")
		return
	}

	if receiver == nil {
		c.Room.Broadcast <- response
		return
	}

	c.Room.Broadcast <- response

	if c.Room != receiver {
		receiver.Broadcast <- response
	}

}
