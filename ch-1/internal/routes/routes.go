// Package routes will help mount and handle routes.
package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
	"github.com/codyonesock/backend_learning/ch-1/internal/users"
)

// RegisterRoutes sets up all the app routes.
func RegisterRoutes(
	r *chi.Mux,
	streamURL string,
	statsService *stats.Service,
	statusService *status.Service,
	userService *users.Service,
) {
	r.Mount("/status", statusService.Handler(statusService, streamURL))

	r.Route("/stats", func(r chi.Router) {
		r.Use(userService.AuthMiddleware)
		r.Get("/", statsService.Handler(statsService).ServeHTTP)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/register", userService.RegisterHandler)
		r.Post("/login", userService.LoginHandler)
	})
}
