package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// User Authentication
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthHandler struct{
	cognito *cognitoidentityprovider.CognitoIdentityProvider
}

func NewAuthHandler(cognito *cognitoidentityprovider.CognitoIdentityProvider) *AuthHandler {
	return &AuthHandler{
		cognito: cognito,
	}
}

// Auth handlers
func (a *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	_, err := a.cognito.SignUp(&cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(os.Getenv("COGNITO_CLIENT_ID")),
		Username: aws.String(user.Email),
		Password: aws.String(user.Password),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(user.Email),
			},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully. Please check your email for verification."})
}

func (a *AuthHandler) ConfirmSignUp(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	_, err := a.cognito.ConfirmSignUp(&cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(os.Getenv("COGNITO_CLIENT_ID")),
		Username:         aws.String(reqData.Email),
		ConfirmationCode: aws.String(reqData.Code),
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email confirmed successfully"})
}

func (a *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	authFlow := "USER_PASSWORD_AUTH"
	result, err := a.cognito.InitiateAuth(&cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: &authFlow,
		ClientId: aws.String(os.Getenv("COGNITO_CLIENT_ID")),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(user.Email),
			"PASSWORD": aws.String(user.Password),
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return
	}

	authResponse := AuthResponse{
		AccessToken:  *result.AuthenticationResult.AccessToken,
		RefreshToken: *result.AuthenticationResult.RefreshToken,
		TokenType:    *result.AuthenticationResult.TokenType,
		ExpiresIn:    int(*result.AuthenticationResult.ExpiresIn),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(authResponse)
}

func (a *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	authFlow := "REFRESH_TOKEN_AUTH"
	result, err := a.cognito.InitiateAuth(&cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: &authFlow,
		ClientId: aws.String(os.Getenv("COGNITO_CLIENT_ID")),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": aws.String(reqData.RefreshToken),
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid refresh token"})
		return
	}

	authResponse := AuthResponse{
		AccessToken: *result.AuthenticationResult.AccessToken,
		TokenType:   *result.AuthenticationResult.TokenType,
		ExpiresIn:   int(*result.AuthenticationResult.ExpiresIn),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(authResponse)
}
