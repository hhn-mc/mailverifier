package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed templates/web/register.html
var registerWebTmpl string

const (
	registerWebTmplName = "register"
)

var webTmpls *template.Template

func init() {
	webTmpls = template.Must(template.New(registerWebTmplName).Parse(registerWebTmpl))
}

func startAPI(cfg config) error {
	emailService := emailService{
		cfg: cfg.Email,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/register", registerGetHandler())
	r.Post("/register", registerPostHandler(emailService, cfg.EmailRegex))
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
	tmpl := webTmpls.Lookup(registerWebTmplName)
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

func registerPostHandler(emailService emailService, emailRegex string) http.HandlerFunc {
	tmpl := webTmpls.Lookup(registerWebTmplName)
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
		if err := emailService.sendVerificationEmail(data, email); err != nil {
			http.Error(w, "Error while sending your email", http.StatusInternalServerError)
			log.Printf("Failed to send email to %q; %s", email, err)
			return
		}

		tmpl.Execute(w, data)
	}
}

func verifyGetHandler() http.HandlerFunc {
	tmpl := webTmpls.Lookup(registerWebTmplName)
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
