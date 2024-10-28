package mailer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/wneessen/go-mail"
)

const port = 587

type GomailMailer struct {
	fromEmail string
	client    *mail.Client
}

func NewGomail(smtpServer string, fromEmail string, password string) (*GomailMailer, error) {
	client, err := mail.NewClient(
		smtpServer,
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithTLSPolicy(mail.TLSMandatory),
		mail.WithUsername(fromEmail),
		mail.WithPassword(password),
		mail.WithPort(port),
	)
	if err != nil {
		return nil, err
	}
	return &GomailMailer{
		fromEmail: fromEmail,
		client:    client,
	}, nil
}

func (gomail *GomailMailer) Send(templateFile string, username string, email string, data any, isSandbox bool) (int, error) {
	if isSandbox {
		return 200, nil
	}

	message := mail.NewMsg()
	if err := message.FromFormat(senderName, gomail.fromEmail); err != nil {
		return -1, fmt.Errorf("failed to set From address: %s", err)
	}
	if err := message.AddToFormat(username, email); err != nil {
		return -1, fmt.Errorf("failed to set To address: %s", err)
	}

	tmpl, err := template.ParseFS(FSys, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return -1, err
	}
	message.Subject(subject.String())

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return -1, err
	}
	message.SetBodyString(mail.TypeTextHTML, body.String())

	message.SetMessageID()
	message.SetDate()
	message.SetBulk()

	for i := 1; i <= maxRetries; i++ {
		if err := gomail.client.DialAndSend(message); err != nil {
			time.Sleep(time.Duration(i*2) * time.Second)
			continue
		}
		return 200, nil
	}
	return -1, err
}
