package csv2table

import (
	"net/smtp"

	"github.com/jordan-wright/email"
)

// Email holds configuration of the email notification
// Currently it supports SMTP with PlainAuth
type Email struct {
	Success bool // send email on success
	Error   bool // send email on error

	SuccessSubject string
	SuccessBody    string
	ErrorSubject   string
	ErrorBody      string

	From string
	To   []string
	Cc   []string
	Bcc  []string

	SMTPServer string // including port e.g. smtp.gmail.com:587

	// config for smtp.PlainAuth
	PlainAuth struct {
		Identity string
		Username string
		Password string
		Host     string
	}
}

const (
	defaultSuccess = true
	defaultError   = true

	defaultSuccessSubject = "csv2table success import"
	defaultSuccessBody    = `
Hello,

csv2table import successfully finished.
	`

	defaultErrorSubject = "csv2table error"
	defaultErrorBody    = `
Hello,

csv2table import failed with the following error:

{error}
	`
)

// emailConfig is the global email configuration
var emailConfig = Email{
	Success:        defaultSuccess,
	Error:          defaultError,
	SuccessSubject: defaultSuccessSubject,
	SuccessBody:    defaultSuccessBody,
	ErrorSubject:   defaultErrorSubject,
	ErrorBody:      defaultErrorBody,
}

// EmailSuccess delivers a success email after an import successfully finished
func EmailSuccess() error {
	email := &email.Email{
		To:      emailConfig.To,
		From:    emailConfig.From,
		Subject: emailConfig.SuccessSubject,
		Text:    []byte(emailConfig.SuccessBody),
	}

	err := email.Send(emailConfig.SMTPServer, smtp.PlainAuth(emailConfig.PlainAuth.Identity, emailConfig.PlainAuth.Username,
		emailConfig.PlainAuth.Password, emailConfig.PlainAuth.Host))

	return err
}
