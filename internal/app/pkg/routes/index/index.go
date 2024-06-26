// Package index contains routes that does not belong to a collection
package index

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router contains all the routes that do not have a collection
func Router(e *env.Env, _ *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Get("/api/doc", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, e.APIDoc, http.StatusMovedPermanently)
	})
	r.Get("/health", health)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/flitlabs/spotoncars_stream", http.StatusTemporaryRedirect)
	})

	return r
}
