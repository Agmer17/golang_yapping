package service

import "gopkg.in/gomail.v2"

type EmailService struct {
	CfgEmail         string
	CfgEmailPassword string
	dialer           *gomail.Dialer
}

func NewEmailService(cfgEmail string, cfgPw string) *EmailService {

	dial := gomail.NewDialer(
		"smtp.gmail.com",
		587,
		cfgEmail,
		cfgPw,
	)

	return &EmailService{
		CfgEmail:         cfgEmail,
		CfgEmailPassword: cfgPw,
		dialer:           dial,
	}

}

func (e *EmailService) SendEmail(to string, subject string, body string) error {

	message := gomail.NewMessage()

	message.SetHeader("From", e.CfgEmail)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)

	message.SetBody("text/html", body)

	return e.dialer.DialAndSend(message)
}
