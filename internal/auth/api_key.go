package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "apikey" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}
