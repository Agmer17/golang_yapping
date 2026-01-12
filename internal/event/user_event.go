package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Agmer17/golang_yapping/pkg"
)

type NewUserEvent struct {
	Email          string
	Username       string
	ActivationLink string
}

func SendVerificationEmail(
	emailSender *pkg.MailSender,
) EventHandler {

	return func(rootCtx context.Context, payload interface{}) {
		eventData := payload.(NewUserEvent)
		eventCtx, eventCancel := context.WithTimeout(rootCtx, 15*time.Second)

		defer eventCancel()

		err := emailSender.SendEmail(eventCtx, eventData.Email, "Verifikasi akun yapping", eventData.ActivationLink)

		if err != nil {
			fmt.Println("ERROR :" + err.Error())
		}
	}

}
