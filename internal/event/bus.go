package event

import (
	"context"
	"sync"

	"github.com/Agmer17/golang_yapping/internal/ws"
)

type EventHandler func(rootCtx context.Context, payload interface{})

type EventBus struct {
	muSubs     sync.RWMutex
	subs       map[string][]EventHandler
	Hub        *ws.Hub
	busContext context.Context
}

func NewEventBus(hub *ws.Hub, eventContext context.Context) *EventBus {
	return &EventBus{
		Hub:        hub,
		busContext: eventContext,
	}

}

func (b *EventBus) Subscribe(eventEndpoint string, f EventHandler) {

	b.muSubs.Lock()
	defer b.muSubs.Unlock()

	b.subs[eventEndpoint] = append(b.subs[eventEndpoint], f)

}

func (b *EventBus) Publish(eventEndpoint string, payload interface{}) {

	b.muSubs.RLock()
	handlers := b.subs[eventEndpoint]
	b.muSubs.RUnlock()

	for _, h := range handlers {

		go h(b.busContext, payload)
	}

}
