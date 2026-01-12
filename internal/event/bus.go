package event

import (
	"context"
	"sync"

	"github.com/Agmer17/golang_yapping/internal/ws"
	"github.com/Agmer17/golang_yapping/pkg"
)

type EventHandler func(rootCtx context.Context, payload interface{})

type EventBus struct {
	muSubs       sync.RWMutex
	subs         map[string][]EventHandler
	Hub          *ws.Hub
	busContext   context.Context
	EmailService *pkg.MailSender
}

func NewEventBus(hub *ws.Hub, eventContext context.Context, emailSender *pkg.MailSender) *EventBus {
	return &EventBus{
		Hub:          hub,
		busContext:   eventContext,
		subs:         make(map[string][]EventHandler),
		EmailService: emailSender,
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
