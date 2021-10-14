package mailer

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/smtp"
)

//go:embed email_verification.html
var verificationEmailFile string

const (
	verificationEmailTmplName = "verification"
)

var emailTmpls *template.Template

func init() {
	emailTmpls = template.Must(template.New(verificationEmailTmplName).Parse(verificationEmailFile))
}

type Service struct {
	Host     string
	SMTPHost string
	Email    string
	Alias    string
	Identity string
	Username string
	Password string
}

func (mail Service) sendEmail(sendTo []string, subject string, msg string) error {
	auth := smtp.PlainAuth(mail.Identity, mail.Username, mail.Password, mail.Host)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg = fmt.Sprintf("To: %s\nFrom: %s <%s>\nSubject: %s\n%s\n%s",
		sendTo, mail.Alias, mail.Email, subject, mime, msg)
	return smtp.SendMail(mail.SMTPHost, auth, mail.Email, sendTo, []byte(msg))
}

type VerificationEmailData struct {
	Code     string
	Username string
	UUID     string
	Time     string
}

func (mail Service) SendVerificationEmail(data VerificationEmailData, sendTo ...string) error {
	subject := "Account Verification"
	tmpl := emailTmpls.Lookup(verificationEmailTmplName)
	w := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}
	return mail.sendEmail(sendTo, subject, w.String())
}
