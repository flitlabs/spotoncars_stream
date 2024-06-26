// Package jobs is a package that contains routes that are related to jobs
package jobs

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router is a function that contains routes that are related to jobs
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsSuperAdmin(h, e, c)
		})
		r.Delete("/end/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
			end(w, r, e, c)
		})
		r.Delete("/reset", func(w http.ResponseWriter, r *http.Request) {
			reset(w, r, e, c)
		})
	})

	return r
}
