package mailer

import "embed"

const (
	senderName          = "Limerence"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed "templates"
var FSys embed.FS

type Client interface {
	Send(templateFile string, username string, email string, data any, isSandbox bool) (int, error)
}
