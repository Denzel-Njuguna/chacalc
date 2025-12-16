package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"chacalc/internal/auth"
	"chacalc/internal/config"
	"chacalc/internal/health"
	authMiddleware "chacalc/internal/middleware"
)

func New(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Handler)

	authHandler := auth.NewHandler(cfg)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.With(authMiddleware.Auth(cfg)).Post("/reset-password", authHandler.ResetPassword)
		r.With(authMiddleware.Auth(cfg)).Post("/update-password", authHandler.UpdatePassword)
		r.With(authMiddleware.Auth(cfg)).Get("/me", authHandler.Me)
	})

	return r
}
