package auth

import (
	"bytes"
	"chacalc/internal/config"
	"encoding/json"
	"net/http"
)

type Handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg: cfg,
	}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	body, _ := json.Marshal(req)

	supabaseReq, err := http.NewRequest(
		http.MethodPost,
		h.cfg.SupabaseURL+"/auth/v1/signup",
		bytes.NewBuffer(body),
	)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	supabaseReq.Header.Set("Content-Type", "application/json")
	supabaseReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)

	client := &http.Client{}
	resp, err := client.Do(supabaseReq)
	if err != nil {
		http.Error(w, "supabase signup failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read Supabase error properly
	if resp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(errBody)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Signup successful.",
	})
}
