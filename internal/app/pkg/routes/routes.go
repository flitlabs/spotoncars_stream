// Package routes contains all the routes
package routes

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/routes/bookings"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/routes/index"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/routes/jobs"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/routes/logs"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/routes/stream"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Route contains various third party connections and eviromental variables that are used throughout the routes
type Route struct {
	E *env.Env
	C *connections.C
}

// Routes contains all the HTTP routes
func (route *Route) Routes() http.Handler {
	r := chi.NewRouter()

	r.Mount("/", index.Router(route.E, route.C))
	r.Mount("/stream", stream.Router(route.E, route.C))
	r.Mount("/bookings", bookings.Router(route.E, route.C))
	r.Mount("/jobs", jobs.Router(route.E, route.C))
	r.Mount("/logs", logs.Router(route.E, route.C))

	return r
}
