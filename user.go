package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/orbulant/chirpy/internal/auth"
	"github.com/orbulant/chirpy/internal/database"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type createUserReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (apiCfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := createUserReqBody{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := apiCfg.dbq.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{User: User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}})
}

func (apiCfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {

	// Get the token from the header
	tokenString, err := auth.GetBearerTokenFromRequest(r)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, apiCfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := createUserReqBody{}
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := apiCfg.dbq.UpdateUserEmailAndPassword(r.Context(), database.UpdateUserEmailAndPasswordParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (apiCfg *apiConfig) handleDeleteAllUsers(w http.ResponseWriter, r *http.Request) {
	_, err := apiCfg.dbq.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete users", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "All users deleted"})
}

type upgradeUserReqBody struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	}
}

func (apiCfg *apiConfig) handleUgradeUserToChirpyRed(w http.ResponseWriter, r *http.Request) {
	// Check the API key
	apiKey, err := auth.GetAPIKey(r)

	if err != nil || apiKey != apiCfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := upgradeUserReqBody{}

	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = apiCfg.dbq.SetIsUserChirpyRedByID(r.Context(), database.SetIsUserChirpyRedByIDParams{
		ID:          params.Data.UserID,
		IsChirpyRed: true,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't upgrade user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
