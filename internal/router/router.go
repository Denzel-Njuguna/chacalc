package router

import (
	"chacalc/internal/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"chacalc/internal/config"
	"chacalc/internal/health"
)

func New(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Handler)

	authHandler := auth.NewHandler(cfg)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
	})

	return r
}
