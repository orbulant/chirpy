package main

import (
	"net/http"
	"os"
)

func (apiCfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {

	platform := os.Getenv("PLATFORM")

	if platform == "dev" {
		apiCfg.handleDeleteAllUsers(w, r)
	}

	apiCfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}
