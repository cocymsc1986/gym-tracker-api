package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// AuthMiddleware is a middleware for authenticating requests using AWS Cognito
type AuthMiddleware struct {
	cognito *cognitoidentityprovider.CognitoIdentityProvider
}

func NewAuthMiddleware(cognito *cognitoidentityprovider.CognitoIdentityProvider) *AuthMiddleware {
	return &AuthMiddleware{
		cognito: cognito,
	}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authorization header required"})
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid authorization format"})
			return
		}

		accessToken := tokenParts[1]

		// Validate token with Cognito
		_, err := m.cognito.GetUser(&cognitoidentityprovider.GetUserInput{
			AccessToken: aws.String(accessToken),
		})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired token"})
			return
		}

		next(w, r)
	}
}
