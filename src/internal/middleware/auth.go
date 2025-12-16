package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"chacalc/src/internal/config"
)

type contextKey string

const userContextKey contextKey = "user"

func Auth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			req, err := http.NewRequest(
				http.MethodGet,
				cfg.SupabaseURL+"/auth/v1/user",
				nil,
			)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			req.Header.Set("apikey", cfg.SupabaseAnonKey)
			req.Header.Set("Authorization", token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode >= 400 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			var user any
			json.NewDecoder(resp.Body).Decode(&user)

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
