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
	template.Must(emailTmpls.New(verificationEmailTmplName).Parse(verificationEmailFile))
}

type emailService struct {
	cfg emailConfig
}

func (mail emailService) sendEmail(sendTo []string, subject string, msg string) error {
	cfg := mail.cfg
	auth := smtp.PlainAuth(cfg.Identity, cfg.Username, cfg.Password, cfg.Host)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg = fmt.Sprintf("To: %s\nFrom: %s\nSubject: %s\n%s\n%s", sendTo, cfg.Email, subject, mime, msg)
	return smtp.SendMail(cfg.SMTPHost, auth, cfg.Email, sendTo, []byte(msg))
}

type verificationEmailData struct {
	Token    string
	Username string
	Time     string
	IP       string
}

func (m emailService) sendVerificationEmail(data verificationEmailData, sendTo ...string) error {
	subject := "HHN Minecraft Verification"
	tmpl := emailTmpls.Lookup(verificationEmailTmplName)
	w := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}
	return m.sendEmail(sendTo, subject, w.String())
}
