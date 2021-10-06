package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/smtp"
)

//go:embed templates/email/verification.html
var verificationEmailFile string

const (
	verificationEmailTmplName = "verification"
)

var emailTmpls *template.Template

func init() {
	emailTmpls = template.Must(template.New(verificationEmailTmplName).Parse(verificationEmailFile))
}

type emailService struct {
	host     string
	smtpHost string
	email    string
	identity string
	username string
	password string
}

func (mail emailService) sendEmail(sendTo []string, subject string, msg string) error {
	auth := smtp.PlainAuth(mail.identity, mail.username, mail.password, mail.host)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg = fmt.Sprintf("To: %s\nFrom: %s\nSubject: %s\n%s\n%s", sendTo, mail.email, subject, mime, msg)
	return smtp.SendMail(mail.smtpHost, auth, mail.email, sendTo, []byte(msg))
}

type verificationEmailData struct {
	Token    string
	Username string
	UUID     string
	Time     string
	IP       string
}

func (mail emailService) sendVerificationEmail(data verificationEmailData, sendTo ...string) error {
	subject := "HHN Minecraft Verification"
	tmpl := emailTmpls.Lookup(verificationEmailTmplName)
	w := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}
	return mail.sendEmail(sendTo, subject, w.String())
}
