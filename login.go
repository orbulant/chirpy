package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/orbulant/chirpy/internal/auth"
	"github.com/orbulant/chirpy/internal/database"
)

type loginReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResBody struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	Token          string `json:"token"`
	RefreshToken   string `json:"refresh_token"`
	ExpirationTime int64  `json:"exp"`
}

func (apiCfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := loginReqBody{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := apiCfg.dbq.GetUserByEmail(r.Context(), params.Email)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	expiry := time.Duration(1) * time.Hour // 1 hour

	token, err := auth.MakeJWT(user.ID, apiCfg.jwtSecret, expiry)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT", err)
		return
	}

	refreshTokenBuffer, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	refreshToken, err := apiCfg.dbq.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTokenBuffer,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, loginResBody{
		ID:             user.ID.String(),
		Email:          user.Email,
		CreatedAt:      user.CreatedAt.String(),
		UpdatedAt:      user.UpdatedAt.String(),
		Token:          token,
		RefreshToken:   refreshToken.Token,
		ExpirationTime: time.Now().Add(expiry).Unix(),
	})
}
