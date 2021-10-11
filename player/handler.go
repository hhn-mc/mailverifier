package player

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gorilla/schema"
)

type DataRepository interface {
	PlayerWithUUIDExists(uuid string) (bool, error)
	PlayerByUUID(uuid string) (Player, error)
	CreatePlayer(player *Player) error

	Verifications(filter VerificationsFilter) ([]Verification, error)
}

func Handler(repo DataRepository) http.Handler {
	r := chi.NewRouter()
	r.Get("/{uuid}", getPlayerHandler(repo))
	r.Post("/", postPlayerHandler(repo))
	r.Get("/{uuid}/verifications")
	return r
}

func getPlayerHandler(repo DataRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if err := validation.Validate(uuid, is.UUIDv4); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		alreadyExists, err := repo.PlayerWithUUIDExists(uuid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if alreadyExists {
			w.WriteHeader(http.StatusConflict)
			return
		}

		player, err := repo.PlayerByUUID(uuid)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			log.Println(err)
			return
		}

		if err := json.NewEncoder(w).Encode(player); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func postPlayerHandler(repo DataRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var player Player
		if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := player.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		alreadyExists, err := repo.PlayerWithUUIDExists(player.UUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if alreadyExists {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if err := repo.CreatePlayer(&player); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if err := json.NewEncoder(w).Encode(player); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getVerifications(repo DataRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var filter VerificationsFilter
		if err := schema.NewDecoder().Decode(&filter, r.URL.Query()); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		_, err := repo.Verifications(filter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println()
		}
	}
}
