package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/orbulant/chirpy/internal/auth"
	"github.com/orbulant/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID     `json:"id"`
	Body      string        `json:"body"`
	UserID    uuid.NullUUID `json:"user_id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type chirpReqBody struct {
	Body   string        `json:"body"`
	UserID uuid.NullUUID `json:"user_id"`
}

var badWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func replaceProfanity(badWords []string, text string) string {
	for _, word := range badWords {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
		text = re.ReplaceAllString(text, "****")
	}
	return text
}

func (apiCfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	// Get JWT from Authorization header
	tokenString, err := auth.GetBearerTokenFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid auth token", err)
		return
	}

	// Validate JWT and extract user ID
	userID, err := auth.ValidateJWT(tokenString, apiCfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid auth token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpReqBody{}
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	} else if len(params.Body) > 400 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", err)
		return
	} else {

		chirp, err := apiCfg.dbq.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   replaceProfanity(badWords, params.Body),
			UserID: uuid.NullUUID{UUID: userID, Valid: true},
		})

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
			return
		}

		respondWithJSON(w, http.StatusCreated, Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		})
	}
}

func (apiCfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorIDStr := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder != "desc" {
		sortOrder = "asc"
	}

	var chirps []database.Chirp
	var err error
	if authorIDStr != "" {
		authorUUID, err := uuid.Parse(authorIDStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author_id", err)
			return
		}
		chirps, err = apiCfg.dbq.GetChirpsByAuthorID(r.Context(), uuid.NullUUID{UUID: authorUUID, Valid: true})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't fetch chirps", err)
			return
		}
	} else {
		chirps, err = apiCfg.dbq.GetAllChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't fetch chirps", err)
		return
	}

	// Sort chirps in-memory based on sortOrder
	sort.Slice(chirps, func(i, j int) bool {
		if sortOrder == "desc" {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})

	var resp []Chirp
	for _, chirp := range chirps {
		resp = append(resp, Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		})
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (apiCfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpID")

	cid, err := uuid.Parse(chirpId)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := apiCfg.dbq.GetChirpByID(r.Context(), cid)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't fetch chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	})
}

func (apiCfg *apiConfig) handleDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpID")

	cid, err := uuid.Parse(chirpId)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	// Check JWT for authentication
	tokenString, err := auth.GetBearerTokenFromRequest(r)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userId, err := auth.ValidateJWT(tokenString, apiCfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	// Fetch the chirp to verify ownership
	chirp, err := apiCfg.dbq.GetChirpByID(r.Context(), cid)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if !chirp.UserID.Valid || chirp.UserID.UUID != userId {
		respondWithError(w, http.StatusForbidden, "You do not have permission to delete this chirp", nil)
		return
	}

	chirp, err = apiCfg.dbq.DeleteChirpByID(r.Context(), cid)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
