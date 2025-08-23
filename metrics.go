package main

import (
	"fmt"
	"net/http"
	"os"
)

func handleMetrics(apiCfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Read the metrics.html file
		htmlBytes, err := os.ReadFile("metrics.html")
		if err != nil {
			http.Error(w, "Could not read metrics.html", http.StatusInternalServerError)
			return
		}
		// Format the HTML with the visit count
		html := fmt.Sprintf(string(htmlBytes), apiCfg.fileserverHits.Load())
		w.Write([]byte(html))
	}
}
