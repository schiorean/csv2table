package csv2table

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// Email holds configuration of the email notification
// Currently it supports SMTP with PlainAuth
type Email struct {
	SendOnSuccess bool // send email on success
	SendOnError   bool // send email on error

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

// EmailTemplateContext is the template context that will be passed to subject and body parser
type EmailTemplateContext struct {
	SuccessCount int
	ErrorCount   int
	Files        []ImportFileStatus
}

const (
	defaultSendOnSuccess = true
	defaultSendOnError   = true
)

// subject templates
// available variables: {{SuccessCount}, {{ErrorCount}}
const (
	defaultSuccessSubject = "csv2table: imported {{.SuccessCount}} file(s)"
	defaultErrorSubject   = "csv2table: import errors"
)

// emial body templates
// available variables: {{SuccessCount}}, {{ErrorCount}, {{Files}}
const (
	defaultSuccessBody = `
Hello,<br/><br/>

{{if .SuccessCount}}
	Successfully imported {{.SuccessCount}} file(s).<br/>
{{end}}	
{{if .ErrorCount}}
	<span style="color:red">{{.ErrorCount}} file(s) produced errors.</span><br/>
{{end}}	

<ol>
{{range .Files}}
	<li>
		{{if .Error}}
			{{.FileName}}: Error: {{.Error}}
		{{else}}
			{{.FileName}}: Imported {{.RowCount}} rows
		{{end}}
	</li>
{{end}}
</ol>

Bye.
`
	defaultErrorBody = defaultSuccessBody
)

// emailConfig is the global email configuration
var emailConfig = Email{
	SendOnSuccess:  defaultSendOnSuccess,
	SendOnError:    defaultSendOnError,
	SuccessSubject: defaultSuccessSubject,
	SuccessBody:    defaultSuccessBody,
	ErrorSubject:   defaultErrorSubject,
	ErrorBody:      defaultErrorBody,
}

// emailConfigured checks if email configuration is present
func emailConfigured() bool {
	return emailConfig.SMTPServer != "" && emailConfig.From != "" && len(emailConfig.To) > 0
}

// createTemplateContext creates the template context used in subject and body parsing
func createTemplateContext(statuses []ImportFileStatus) EmailTemplateContext {

	var successCount, errorCount int
	var status ImportFileStatus

	for _, status = range statuses {
		if status.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	return EmailTemplateContext{
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		Files:        statuses,
	}
}

// parseTemplate parses an email template (subject or body)
func parseTemplate(tpl string, context EmailTemplateContext) (string, error) {
	// create subject
	t, err := template.New("email").Parse(tpl)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	err = t.Execute(&output, context)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

// sendEmailSuccess delivers a success email after an import successfully finished
func sendEmailSuccess(statuses []ImportFileStatus) error {
	e := &email.Email{
		To:      emailConfig.To,
		From:    emailConfig.From,
		Subject: emailConfig.SuccessSubject,
		Text:    []byte(emailConfig.SuccessBody),
	}

	var err error

	// create subject
	context := createTemplateContext(statuses)
	e.Subject, err = parseTemplate(emailConfig.SuccessSubject, context)
	if err != nil {
		return err
	}

	text, err := parseTemplate(emailConfig.SuccessBody, context)
	if err != nil {
		return err
	}
	e.HTML = []byte(text)

	// fmt.Println(e.Subject)
	// fmt.Println(string(e.HTML))
	// return nil

	err = e.Send(emailConfig.SMTPServer, smtp.PlainAuth(emailConfig.PlainAuth.Identity, emailConfig.PlainAuth.Username,
		emailConfig.PlainAuth.Password, emailConfig.PlainAuth.Host))

	return err
}

// sendEmailError delivers a success email after an import successfully finished
func sendEmailError(statuses []ImportFileStatus) error {
	e := &email.Email{
		To:      emailConfig.To,
		From:    emailConfig.From,
		Subject: emailConfig.SuccessSubject,
		Text:    []byte(emailConfig.SuccessBody),
	}

	var err error

	// create subject
	context := createTemplateContext(statuses)
	e.Subject, err = parseTemplate(emailConfig.ErrorSubject, context)
	if err != nil {
		return err
	}

	text, err := parseTemplate(emailConfig.ErrorBody, context)
	if err != nil {
		return err
	}
	e.HTML = []byte(text)

	// fmt.Println(e.Subject)
	// fmt.Println(string(e.HTML))
	// return nil

	err = e.Send(emailConfig.SMTPServer, smtp.PlainAuth(emailConfig.PlainAuth.Identity, emailConfig.PlainAuth.Username,
		emailConfig.PlainAuth.Password, emailConfig.PlainAuth.Host))

	return err
}
