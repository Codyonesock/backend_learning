// Package routes will help mount and handle routes.
package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

// RegisterRoutes sets up all the app routes.
func RegisterRoutes(r *chi.Mux, statsService *stats.Service, statusService *status.Service, streamURL string) {
	r.Mount("/status", StatusRoutes(statusService, streamURL))
	r.Mount("/stats", StatsRoutes(statsService))
}

// StatusRoutes defines the /status routes.
func StatusRoutes(statusService *status.Service, streamURL string) http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		if err := statusService.ProcessStream(streamURL); err != nil {
			http.Error(w, "Error processing stream", http.StatusInternalServerError)
		}
	})

	return r
}

// StatsRoutes defines the /stats routes.
func StatsRoutes(statsService *stats.Service) http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		if err := statsService.GetStats(w); err != nil {
			http.Error(w, "Error getting stats", http.StatusInternalServerError)
		}
	})

	return r
}
