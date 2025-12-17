package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func GetUserEmailFromToken(bearerToken string) (string, error) {
	// 1. Remove "Bearer " prefix if present
	tokenParts := strings.Split(bearerToken, " ")
	if len(tokenParts) != 2 {
		return "", errors.New("invalid authorization header")
	}
	tokenStr := tokenParts[1]

	// 2. Parse JWT (we won't verify signature here for simplicity)
	token, _, err := jwt.NewParser().ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	// 3. Extract email
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if email, ok := claims["email"].(string); ok {
			return email, nil
		}
	}

	return "", errors.New("email not found in token")
}

func extractUserIDFromJWT(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return ""
	}

	tokenStr := parts[1]
	tokenParts := strings.Split(tokenStr, ".")
	if len(tokenParts) < 2 {
		return ""
	}

	payload, _ := base64.RawStdEncoding.DecodeString(tokenParts[1])

	var claims map[string]any
	json.Unmarshal(payload, &claims)

	sub, _ := claims["sub"].(string)
	return sub
}
