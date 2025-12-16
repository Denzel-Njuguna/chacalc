package auth

import (
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
