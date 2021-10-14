package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/hhn-mc/mailverifier/db"
)

type ctxKey int

const (
	CtxPlayerUUIDKey ctxKey = iota
	CtxPlayerUsernameKey
)

func loadPlayer(db db.DB) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uuid := chi.URLParam(r, "uuid")
			if err := validation.Validate(uuid, is.UUIDv4); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			alreadyExists, err := db.PlayerWithUUIDExists(uuid)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !alreadyExists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			player, err := db.PlayerByUUID(uuid)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				return
			}

			ctx := context.WithValue(r.Context(), CtxPlayerUUIDKey, player.UUID)
			ctx = context.WithValue(ctx, CtxPlayerUsernameKey, player.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
