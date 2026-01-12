package event

const NewUserCreated = "user.created"

func SetupEvent(bus *EventBus) {
	bus.Subscribe(NewUserCreated, SendVerificationEmail(bus.EmailService))
}
