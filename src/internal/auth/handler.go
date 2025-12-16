package auth

import (
	"bytes"
	"chacalc/src/internal/config"
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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	body, _ := json.Marshal(req)

	supabaseReq, err := http.NewRequest(
		http.MethodPost,
		h.cfg.SupabaseURL+"/auth/v1/token?grant_type=password",
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
		http.Error(w, "supabase login failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(errBody)
		return
	}

	var loginResp LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	supabaseReq, err := http.NewRequest(
		http.MethodPost,
		h.cfg.SupabaseURL+"/auth/v1/logout",
		nil,
	)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	supabaseReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	supabaseReq.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(supabaseReq)
	if err != nil {
		http.Error(w, "supabase logout failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	body, _ := json.Marshal(req)

	supabaseReq, err := http.NewRequest(
		http.MethodPost,
		h.cfg.SupabaseURL+"/auth/v1/recover",
		bytes.NewBuffer(body),
	)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	supabaseReq.Header.Set("Content-Type", "application/json")
	supabaseReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)

	client := &http.Client{}
	_, _ = client.Do(supabaseReq)
	// intentionally ignoring response for security

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If an account exists, a reset email has been sent",
	})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// 1. Get token from Authorization header
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	// 2. Decode request body
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	// 3. Prepare Supabase request body
	body, _ := json.Marshal(map[string]string{
		"password": req.Password,
	})

	// 4. Create Supabase request
	supabaseReq, err := http.NewRequest(
		http.MethodPut,
		h.cfg.SupabaseURL+"/auth/v1/user",
		bytes.NewBuffer(body),
	)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	supabaseReq.Header.Set("Content-Type", "application/json")
	supabaseReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	supabaseReq.Header.Set("Authorization", token)

	// 5. Execute request
	client := &http.Client{}
	resp, err := client.Do(supabaseReq)
	if err != nil {
		http.Error(w, "supabase reset password failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 6. Handle Supabase errors
	if resp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(errBody)
		return
	}

	// 7. Success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successful. Please log in again.",
	})
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "new password is required", http.StatusBadRequest)
		return
	}

	body, _ := json.Marshal(map[string]string{
		"password": req.NewPassword,
	})

	supabaseReq, err := http.NewRequest(
		http.MethodPut,
		h.cfg.SupabaseURL+"/auth/v1/user",
		bytes.NewBuffer(body),
	)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	supabaseReq.Header.Set("Content-Type", "application/json")
	supabaseReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	supabaseReq.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(supabaseReq)
	if err != nil {
		http.Error(w, "failed to update password", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(errBody)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password updated successfully. Please log in again.",
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	// 1. Get token from Authorization header
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	// 2. Create Supabase request to get current user
	supabaseReq, err := http.NewRequest(
		http.MethodGet,
		h.cfg.SupabaseURL+"/auth/v1/user",
		nil,
	)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	// 3. Set headers
	supabaseReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	supabaseReq.Header.Set("Authorization", token)

	// 4. Execute request
	client := &http.Client{}
	resp, err := client.Do(supabaseReq)
	if err != nil {
		http.Error(w, "supabase get user failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 5. Decode response
	var user any
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		http.Error(w, "failed to decode response", http.StatusInternalServerError)
		return
	}

	// 6. Handle Supabase errors
	if resp.StatusCode >= 400 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(user)
		return
	}

	// 7. Success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
