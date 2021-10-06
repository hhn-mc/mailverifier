package main

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gofrs/uuid"
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

type api struct {
	bind       string
	mailer     emailService
	db         database
	emailRegex string
	creds      map[string]string
}

func (api api) listenAndServe() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/register/{token}", registerGetHandler())
	r.Post("/register", registerPostHandler(api.mailer, api.emailRegex))
	r.Get("/verify/{token}", verifyGetHandler())
	r.Mount("/admin", adminRouter(api.creds))

	srv := http.Server{
		Addr:    api.bind,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func adminRouter(creds map[string]string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.BasicAuth("", creds))
	r.Post("/players", playersPostHandler())
	return r
}

type playerPostJSON struct {
	UUID string `json:"uuid"`
}

func playersPostHandler(db database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var jsonData playerPostJSON
		if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
			http.Error(w, "Failed to decode json", http.StatusBadRequest)
			return
		}

		token, err := uuid.NewV4()
		if err != nil {
			http.Error(w, "Failed to create UUIDv4", http.StatusInternalServerError)
			log.Printf("Failed to create UUIDv4; %s", err)
			return
		}

		tx, err := db.Begin()
		tx.Query()

		if err := db.CreatePlayer(jsonData.UUID); err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Fatalf("Failed to create player in database; %s", err)
			return
		}

		tokenString := token.String()
		if err := db.CreatePlayerEmailVerification(tokenString, jsonData.UUID); err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Fatalf("Failed to create verification; %s", err)
			return
		}


	}
}

type registerWebData struct {
	Token    string
	Username string
}

func registerGetHandler() http.HandlerFunc {
	tmpl := webTmpls.Lookup(registerWebTmplName)
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")

		data := registerWebData{
			Token:    token,
			Username: "",
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
		data := struct {
			Username string
		}{
			Username: "",
		}

		tmpl.Execute(w, data)
	}
}
