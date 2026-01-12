package pkg

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/wneessen/go-mail"
)

type MailSender struct {
	client *mail.Client
	from   string
}

func NewMailSender(email, appPassword string) (*MailSender, error) {
	c, err := mail.NewClient(
		"smtp.gmail.com",
		mail.WithPort(587),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(email),
		mail.WithPassword(appPassword),
		mail.WithTLSConfig(&tls.Config{
			ServerName: "smtp.gmail.com",
		}),
		mail.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}

	return &MailSender{
		client: c,
		from:   email,
	}, nil

}

func (ms *MailSender) SendEmail(ctx context.Context, to, subject, body string) error {

	msg := mail.NewMsg()

	if err := msg.From(ms.from); err != nil {
		return err
	}
	if err := msg.To(to); err != nil {
		return err
	}

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)

	return ms.client.DialAndSendWithContext(ctx, msg)

}
