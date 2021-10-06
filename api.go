package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"text/template"

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
	r.Post("/register", registerPostHandler(cfg.EmailRegex))
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

func registerPostHandler(emailRegex string) http.HandlerFunc {
	tmpl := template.Must(template.New("verifyEmail").Parse(verifyEmailTemplate))
	emailPattern := regexp.MustCompile(emailRegex)
	return func(w http.ResponseWriter, r *http.Request) {
		email := fromURLValues(r.Form, "email")
		if !emailPattern.MatchString(email) {
			http.Error(w, "Email in invalid format", http.StatusBadRequest)
			return
		}

		query := r.URL.Query()
		data := struct {
			Username string
		}{
			Username: fromURLValues(query, "username"),
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

func sendVerificationEmail(cfg config, sendTo string) error {
	auth := smtp.PlainAuth(cfg.Email.Identity, cfg.Email.Username, cfg.Email.Password, cfg.Email.Host)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	to := []string{sendTo}
	subject := "HHN Minecraft Verify"
	msg := fmt.Sprintf("To: %s\nFrom: %s\nSubject: %s\n%s\n%s", sendTo, cfg.Email.Identity, subject, mime, verifyEmailTemplate)
	return smtp.SendMail(cfg.Email.SMTPHost, auth, cfg.Email.Email, to, []byte(msg))
}
