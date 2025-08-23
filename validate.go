package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
)

type parameters struct {
	// these tags indicate how the keys in the JSON should be mapped to the struct fields
	// the struct fields must be exported (start with a capital letter) if you want them parsed
	Body string `json:"body"`
}

type responseBody struct {
	CleanedBody string `json:"cleaned_body"`
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

func handleValidateBodyLength(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong", err)
		return
	} else if len(params.Body) > 400 {
		respondWithError(w, 400, "Chirp is too long", err)
		return
	} else {

		respBody := responseBody{
			CleanedBody: replaceProfanity(badWords, params.Body),
		}

		respondWithJSON(w, 200, respBody)
	}
}
