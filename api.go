package main

import (
	_ "embed"
	"encoding/base64"
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
	db         *database
	emailRegex string
	creds      map[string]string
}

func (api api) listenAndServe() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/register/{token}", registerGetHandler(api.db))
	r.Post("/register", registerPostHandler(api.db, api.mailer, api.emailRegex))
	r.Get("/verify/{token}", verifyGetHandler())
	r.Mount("/auth", api.adminRouter(api.creds))

	srv := http.Server{
		Addr:    api.bind,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func (api *api) adminRouter(creds map[string]string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.BasicAuth("", creds))
	r.Post("/players/verifications", verificationsPostHandler(api.db))
	return r
}

type playerPostJSONReq struct {
	Username string `json:"username"`
	UUID     string `json:"uuid"`
}

type playerPostJSONResp struct {
	Token string `json:"token"`
}

func verificationsPostHandler(db *database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var jsonReq playerPostJSONReq
		if err := json.NewDecoder(r.Body).Decode(&jsonReq); err != nil {
			http.Error(w, "Failed to decode json", http.StatusBadRequest)
			return
		}

		token, err := db.getActiveVerificationTokenForPlayerUUID(jsonReq.UUID)
		if err != nil {
			token, err = uuid.NewV4()
			if err != nil {
				http.Error(w, "Failed to create UUIDv4", http.StatusInternalServerError)
				log.Printf("Failed to create UUIDv4; %s", err)
				return
			}

			if err := db.createPlayerIfNotExistAndStartVerification(
				jsonReq.Username, jsonReq.UUID, token.String()); err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				log.Printf("Failed to create player or verification in database; %s", err)
				return
			}
		}

		if err := json.NewEncoder(w).Encode(playerPostJSONResp{
			Token: base64.URLEncoding.EncodeToString(token.Bytes()),
		}); err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Printf("Failed to marshal json; %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

type registerWebData struct {
	Token string
	UUID  string
}

func registerGetHandler(db *database) http.HandlerFunc {
	tmpl := webTmpls.Lookup(registerWebTmplName)
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token == "" {
			http.Error(w, "Missing token", http.StatusBadRequest)
			return
		}

		bb, err := base64.URLEncoding.DecodeString(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusBadRequest)
			return
		}

		uuid, err := db.getPlayerUUIDFromToken(string(bb))
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			log.Printf("Failed to query player uuid from database; %s", err)
			return
		}

		data := registerWebData{
			Token: token,
			UUID:  uuid.String(),
		}

		tmpl.Execute(w, data)
	}
}

func registerPostHandler(db *database, emailService emailService, emailRegex string) http.HandlerFunc {
	tmpl := webTmpls.Lookup(registerWebTmplName)
	emailPattern := regexp.MustCompile(emailRegex)
	return func(w http.ResponseWriter, r *http.Request) {
		encodedToken := r.FormValue("token")
		if encodedToken == "" {
			http.Error(w, "Missing token", http.StatusBadRequest)
			return
		}

		playerUUID := r.FormValue("uuid")
		if playerUUID == "" {
			http.Error(w, "Missing UUID", http.StatusBadRequest)
			return
		}

		bb, err := base64.URLEncoding.DecodeString(encodedToken)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusBadRequest)
			return
		}
		token := string(bb)

		tokenExists, err := db.doesVerificationTokenExist(token, playerUUID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			log.Printf("Failed to verify token in database; %s", err)
			return
		}

		if !tokenExists {
			http.Error(w, "Invalid token", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		if email == "" || !emailPattern.MatchString(email) {
			http.Error(w, "Email in invalid format", http.StatusBadRequest)
			log.Printf("Email did not match: %q", email)
			return
		}

		username, err := db.getPlayerUsernameFromToken(token)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			log.Printf("Failed to query player username from database; %s", err)
			return
		}

		data := verificationEmailData{
			Token:    "", // TODO: Generate email token
			UUID:     playerUUID,
			Username: username,
			Time:     time.Now().String(),
			IP:       r.RemoteAddr,
		}
		if err := emailService.sendVerificationEmail(data, email); err != nil {
			http.Error(w, "Error while sending your email", http.StatusInternalServerError)
			log.Printf("Failed to send email to %q; %s", email, err)
			return
		}
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
