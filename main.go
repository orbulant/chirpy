package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	indexFile := http.FileServer(http.Dir("."))

	mux.Handle("/", indexFile)
	mux.Handle("/app", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/assets", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets"))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// The line below is not needed as it automatically does it on first .Write
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
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
