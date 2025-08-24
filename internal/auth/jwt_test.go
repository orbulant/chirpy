package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT_Success(t *testing.T) {
	secret := "supersecret"
	userID := uuid.New()
	token, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	parsedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if parsedID != userID {
		t.Errorf("Expected userID %v, got %v", userID, parsedID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	secret := "supersecret"
	userID := uuid.New()
	token, err := MakeJWT(userID, secret, -time.Minute) // Already expired
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	secret := "supersecret"
	wrongSecret := "wrongsecret"
	userID := uuid.New()
	token, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Error("Expected error for wrong secret, got nil")
	}
}

func TestGetBearerTokenFromRequest_Success(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer SOME_TOKEN")

	token, err := GetBearerTokenFromRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if token != "SOME_TOKEN" {
		t.Errorf("Expected token 'SOME_TOKEN', got '%s'", token)
	}
}
