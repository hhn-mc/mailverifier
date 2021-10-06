package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed templates/email/verify.html
var verifyEmailTemplate string

//go:embed templates/web/register.html
var registerWebTemplate string

func startAPI(cfg config) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/register", registerGetHandler())
	r.Post("/register", registerPostHandler(cfg.Email, cfg.EmailRegex))
	r.Post("/verify", verifyGetHandler())

	srv := http.Server{
		Addr:    cfg.API.Bind,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func fromURLValues(query url.Values, key string) string {
	v, ok := query[key]
	if !ok {
		return ""
	}
	return v[0]
}

func registerGetHandler() http.HandlerFunc {
	tmpl := template.Must(template.New("registerWeb").Parse(registerWebTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		data := struct {
			Username string
		}{
			Username: fromURLValues(query, "username"),
		}

		tmpl.Execute(w, data)
	}
}

func registerPostHandler(emailCfg emailConfig, emailRegex string) http.HandlerFunc {
	tmpl := template.Must(template.New("verifyEmail").Parse(verifyEmailTemplate))
	emailPattern := regexp.MustCompile(emailRegex)
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		if !emailPattern.MatchString(email) {
			http.Error(w, "Email in invalid format", http.StatusBadRequest)
			log.Printf("Email did not match: %q", email)
			return
		}

		data := verificationEmailData{
			Token:    "",
			Username: r.FormValue("username"),
			Time:     time.Now().String(),
			IP:       r.RemoteAddr,
		}
		if err := sendVerificationEmail(emailCfg, email, data); err != nil {
			http.Error(w, "Error while sending your email", http.StatusInternalServerError)
			log.Printf("Failed to send email to %q; %s", email, err)
			return
		}

		tmpl.Execute(w, data)
	}
}

func verifyGetHandler() http.HandlerFunc {
	tmpl := template.Must(template.New("registerEmail").Parse(verifyEmailTemplate))
	return func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		data := struct {
			Username string
		}{
			Username: fromURLValues(query, "username"),
		}

		tmpl.Execute(w, data)
	}
}

type verificationEmailData struct {
	Token    string
	Username string
	Time     string
	IP       string
}

func sendVerificationEmail(cfg emailConfig, sendTo string, data verificationEmailData) error {
	tmpl := template.Must(template.New("verificationEmail").Parse(verifyEmailTemplate))
	auth := smtp.PlainAuth(cfg.Identity, cfg.Username, cfg.Password, cfg.Host)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	to := []string{sendTo}
	subject := "HHN Minecraft Verify"
	msg := fmt.Sprintf("To: %s\nFrom: %s\nSubject: %s\n%s\n", sendTo, cfg.Identity, subject, mime)
	w := bytes.NewBuffer([]byte(msg))
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}
	return smtp.SendMail(cfg.SMTPHost, auth, cfg.Email, to, w.Bytes())
}
