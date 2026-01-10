package userevent

type NewUserEvent struct {
	Email          string
	Username       string
	ActivationLink string
}
