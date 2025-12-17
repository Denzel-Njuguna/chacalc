package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Request/Response structs
type CreateChamaRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CreateChamaResponse struct {
	Message    string `json:"message"`
	ChamaID    string `json:"chama_id"`
	InviteCode string `json:"invite_code"`
}

type JoinChamaRequest struct {
	InviteCode string `json:"invite_code"`
}

type JoinChamaResponse struct {
	Message string `json:"message"`
	ChamaID string `json:"chama_id"`
}

type AddMemberRequest struct {
	ChamaID string `json:"chama_id"`
	UserID  string `json:"user_id"`
	Role    string `json:"role,omitempty"` // defaults to "member"
}

type AddMemberResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	ChamaID string `json:"chama_id"`
}

// Generate a unique 8-character invite code
func generateInviteCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// CreateChama creates a new organization and assigns the creator as admin
func (h *Handler) CreateChama(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	userID := extractUserIDFromJWT(token)
	if userID == "" {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var req CreateChamaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "chama name is required", http.StatusBadRequest)
		return
	}

	// Generate unique invite code
	inviteCode := generateInviteCode()

	// FIXED: Use snake_case column names to match database schema
	body, _ := json.Marshal(map[string]any{
		"name":           req.Name,
		"description":    req.Description,
		"member_count":   1,
		"chama_holdings": 0,
		"invite_code":    inviteCode,
		"created_at":     time.Now().Format(time.RFC3339),
	})

	createReq, _ := http.NewRequest(
		http.MethodPost,
		h.cfg.SupabaseURL+"/rest/v1/chamas",
		bytes.NewBuffer(body),
	)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	createReq.Header.Set("Authorization", token)
	createReq.Header.Set("Prefer", "return=representation")

	client := &http.Client{}
	resp, err := client.Do(createReq)
	if err != nil {
		http.Error(w, "failed to create chama", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(errBody)
		return
	}

	var created []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil || len(created) == 0 {
		http.Error(w, "failed to parse chama response", http.StatusInternalServerError)
		return
	}

	// FIXED: Use snake_case column name
	chamaID := created[0]["chama_id"].(string)

	// Add creator as admin
	// FIXED: Use snake_case column names
	memberBody, _ := json.Marshal(map[string]any{
		"user_id":   userID,
		"chama_id":  chamaID,
		"role":      "admin",
		"join_date": time.Now().Format(time.RFC3339),
	})

	memberReq, _ := http.NewRequest(
		http.MethodPost,
		h.cfg.SupabaseURL+"/rest/v1/chamamembers",
		bytes.NewBuffer(memberBody),
	)
	memberReq.Header.Set("Content-Type", "application/json")
	memberReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	memberReq.Header.Set("Authorization", token)

	resp2, err := client.Do(memberReq)
	if err != nil || resp2.StatusCode >= 400 {
		http.Error(w, "failed to add creator to chama", http.StatusBadGateway)
		return
	}
	defer resp2.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateChamaResponse{
		Message:    "Chama created successfully",
		ChamaID:    chamaID,
		InviteCode: inviteCode,
	})
}

// JoinChama allows a user to join an organization using an invite code
func (h *Handler) JoinChama(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	userID := extractUserIDFromJWT(token)
	if userID == "" {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var req JoinChamaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.InviteCode == "" {
		http.Error(w, "invite code is required", http.StatusBadRequest)
		return
	}

	client := &http.Client{}

	// Find Chama by invite code
	chamaReq, _ := http.NewRequest(
		http.MethodGet,
		h.cfg.SupabaseURL+"/rest/v1/chamas?invite_code=eq."+req.InviteCode,
		nil,
	)
	chamaReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	chamaReq.Header.Set("Authorization", token)

	resp, err := client.Do(chamaReq)
	if err != nil || resp.StatusCode >= 400 {
		http.Error(w, "failed to fetch chama", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var chamas []map[string]any
	json.NewDecoder(resp.Body).Decode(&chamas)
	if len(chamas) == 0 {
		http.Error(w, "invalid invite code", http.StatusBadRequest)
		return
	}

	// FIXED: Use snake_case column name
	chamaID := chamas[0]["chama_id"].(string)

	// Check if user is already a member
	// FIXED: Use snake_case column names in query
	checkReq, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/rest/v1/chamamembers?user_id=eq.%s&chama_id=eq.%s",
			h.cfg.SupabaseURL, userID, chamaID),
		nil,
	)
	checkReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	checkReq.Header.Set("Authorization", token)

	checkResp, err := client.Do(checkReq)
	if err == nil {
		defer checkResp.Body.Close()
		var existing []map[string]any
		json.NewDecoder(checkResp.Body).Decode(&existing)
		if len(existing) > 0 {
			http.Error(w, "you are already a member of this chama", http.StatusBadRequest)
			return
		}
	}

	// Add user to chama members
	// FIXED: Use snake_case column names
	memberBody, _ := json.Marshal(map[string]any{
		"user_id":   userID,
		"chama_id":  chamaID,
		"role":      "member",
		"join_date": time.Now().Format(time.RFC3339),
	})

	memberReq, _ := http.NewRequest(
		http.MethodPost,
		h.cfg.SupabaseURL+"/rest/v1/chamamembers",
		bytes.NewBuffer(memberBody),
	)
	memberReq.Header.Set("Content-Type", "application/json")
	memberReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	memberReq.Header.Set("Authorization", token)

	resp2, err := client.Do(memberReq)
	if err != nil || resp2.StatusCode >= 400 {
		http.Error(w, "failed to join chama", http.StatusBadGateway)
		return
	}
	defer resp2.Body.Close()

	// Update chama member count
	currentCount := int(chamas[0]["member_count"].(float64))
	updateBody, _ := json.Marshal(map[string]any{
		"member_count": currentCount + 1,
	})
	// FIXED: Use snake_case column name in query
	updateReq, _ := http.NewRequest(
		http.MethodPatch,
		h.cfg.SupabaseURL+"/rest/v1/chamas?chama_id=eq."+chamaID,
		bytes.NewBuffer(updateBody),
	)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	updateReq.Header.Set("Authorization", token)

	client.Do(updateReq)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JoinChamaResponse{
		Message: "Joined Chama successfully",
		ChamaID: chamaID,
	})
}

// AddMemberToChama allows an admin to directly add a user to their organization
func (h *Handler) AddMemberToChama(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	adminUserID := extractUserIDFromJWT(token)
	if adminUserID == "" {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ChamaID == "" || req.UserID == "" {
		http.Error(w, "chama_id and user_id are required", http.StatusBadRequest)
		return
	}

	// Default role to "member" if not specified
	role := req.Role
	if role == "" {
		role = "member"
	}

	client := &http.Client{}

	// Verify the requester is an admin of this chama
	// FIXED: Use snake_case column names in query
	adminCheckReq, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/rest/v1/chamamembers?user_id=eq.%s&chama_id=eq.%s&role=eq.admin",
			h.cfg.SupabaseURL, adminUserID, req.ChamaID),
		nil,
	)
	adminCheckReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	adminCheckReq.Header.Set("Authorization", token)

	adminResp, err := client.Do(adminCheckReq)
	if err != nil || adminResp.StatusCode >= 400 {
		http.Error(w, "failed to verify admin status", http.StatusBadGateway)
		return
	}
	defer adminResp.Body.Close()

	var admins []map[string]any
	json.NewDecoder(adminResp.Body).Decode(&admins)
	if len(admins) == 0 {
		http.Error(w, "only admins can add members", http.StatusForbidden)
		return
	}

	// Check if user is already a member
	// FIXED: Use snake_case column names in query
	checkReq, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/rest/v1/chamamembers?user_id=eq.%s&chama_id=eq.%s",
			h.cfg.SupabaseURL, req.UserID, req.ChamaID),
		nil,
	)
	checkReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	checkReq.Header.Set("Authorization", token)

	checkResp, err := client.Do(checkReq)
	if err == nil {
		defer checkResp.Body.Close()
		var existing []map[string]any
		json.NewDecoder(checkResp.Body).Decode(&existing)
		if len(existing) > 0 {
			http.Error(w, "user is already a member of this chama", http.StatusBadRequest)
			return
		}
	}

	// Add user to chama
	// FIXED: Use snake_case column names
	memberBody, _ := json.Marshal(map[string]any{
		"user_id":   req.UserID,
		"chama_id":  req.ChamaID,
		"role":      role,
		"join_date": time.Now().Format(time.RFC3339),
	})

	memberReq, _ := http.NewRequest(
		http.MethodPost,
		h.cfg.SupabaseURL+"/rest/v1/chamamembers",
		bytes.NewBuffer(memberBody),
	)
	memberReq.Header.Set("Content-Type", "application/json")
	memberReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	memberReq.Header.Set("Authorization", token)

	resp, err := client.Do(memberReq)
	if err != nil || resp.StatusCode >= 400 {
		http.Error(w, "failed to add member to chama", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Update member count
	// FIXED: Use snake_case column name in query
	getChamaReq, _ := http.NewRequest(
		http.MethodGet,
		h.cfg.SupabaseURL+"/rest/v1/chamas?chama_id=eq."+req.ChamaID,
		nil,
	)
	getChamaReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	getChamaReq.Header.Set("Authorization", token)

	chamaResp, err := client.Do(getChamaReq)
	if err == nil {
		defer chamaResp.Body.Close()
		var chamas []map[string]any
		json.NewDecoder(chamaResp.Body).Decode(&chamas)
		if len(chamas) > 0 {
			currentCount := int(chamas[0]["member_count"].(float64))
			updateBody, _ := json.Marshal(map[string]any{
				"member_count": currentCount + 1,
			})
			// FIXED: Use snake_case column name in query
			updateReq, _ := http.NewRequest(
				http.MethodPatch,
				h.cfg.SupabaseURL+"/rest/v1/chamas?chama_id=eq."+req.ChamaID,
				bytes.NewBuffer(updateBody),
			)
			updateReq.Header.Set("Content-Type", "application/json")
			updateReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
			updateReq.Header.Set("Authorization", token)
			client.Do(updateReq)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AddMemberResponse{
		Message: "Member added successfully",
		UserID:  req.UserID,
		ChamaID: req.ChamaID,
	})
}
