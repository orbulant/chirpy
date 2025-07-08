package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	indexFile := http.FileServer(http.Dir("."))
	apiCfg := apiConfig{}

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/", indexFile)
	mux.Handle("/app", apiCfg.middlewareMetricsInc(handler))
	mux.Handle("/app/assets", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets"))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// The line below is not needed as it automatically does it on first .Write
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		buf := []byte{}
		buf = fmt.Appendf(buf, "Hits: %d", apiCfg.fileserverHits.Load())
		w.Write(buf)
	})
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileserverHits.Store(0)
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Server started on port 8080")

	err := server.ListenAndServe()

	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
