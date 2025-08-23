package main

import (
	"net/http"
	"os"
)

func handleReset(apiCfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		platform := os.Getenv("PLATFORM")

		if platform == "dev" {
			handleDeleteAllUsers(apiCfg)
		}

		apiCfg.fileserverHits.Store(0)
		w.WriteHeader(http.StatusOK)
	}
}
