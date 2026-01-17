package event

const NewUserCreated = "user.created"
const WsEventSendPayload = "ws.send.payload"

func SetupEvent(bus *EventBus) {
	bus.Subscribe(NewUserCreated, SendVerificationEmail(bus.EmailService))
	bus.Subscribe(WsEventSendPayload, sendPayload(bus.Hub))
}
