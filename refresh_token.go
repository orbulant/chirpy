package main

import (
	"net/http"
	"time"

	"github.com/orbulant/chirpy/internal/auth"
	"github.com/orbulant/chirpy/internal/database"
)

type refreshResBody struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Expiration   int64  `json:"exp"`
}

func (apiCfg *apiConfig) handleRefreshRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get the Authorization token
	token, err := auth.GetBearerTokenFromRequest(r)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or malformed token", err)
		return
	}

	// Get the refresh token from the database
	dbToken, err := apiCfg.dbq.GetRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Check if the token is expired
	if dbToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired", err)
		return
	}

	// Check if the token has been revoked
	if dbToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked", nil)
		return
	}

	// Create new Refresh Token
	newExpiry := time.Now().Add(60 * 24 * time.Hour) // 60 days
	newRefreshTokenValue, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	// Save the new refresh token to the database
	newRefreshToken, err := apiCfg.dbq.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     newRefreshTokenValue,
		UserID:    dbToken.UserID,
		ExpiresAt: newExpiry,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	accessTokenExpiry := time.Hour
	accessToken, err := auth.MakeJWT(dbToken.UserID, apiCfg.jwtSecret, accessTokenExpiry)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, refreshResBody{
		Token:        accessToken,
		RefreshToken: newRefreshToken.Token,
		Expiration:   newRefreshToken.ExpiresAt.Unix(),
	})
}

func (apiCfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	// Get the Authorization token
	token, err := auth.GetBearerTokenFromRequest(r)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or malformed token", err)
		return
	}

	// Get the refresh token from the database
	dbToken, err := apiCfg.dbq.GetRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Revoke the token by setting revoked_at to now
	err = apiCfg.dbq.RevokeRefreshToken(r.Context(), dbToken.Token)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
