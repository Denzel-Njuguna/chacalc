package auth

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SupabaseAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         any    `json:"user"`
}
