package event

import (
	"context"
	"encoding/json"

	"github.com/Agmer17/golang_yapping/internal/ws"
)

func sendPayload(
	hub *ws.Hub,
) EventHandler {

	return func(rootCtx context.Context, payload interface{}) {

		data := payload.(ws.WebsocketEvent)
		wsEventByte, _ := json.Marshal(data)

		hub.SendPayloadTo(data.Receiver, wsEventByte)

	}

}
