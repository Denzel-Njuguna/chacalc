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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         any    `json:"user"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Password string `json:"password"`
}

type UpdatePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type OnboardingRequest struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

type OnboardingResponse struct {
	Message string `json:"message"`
}
