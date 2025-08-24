package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/orbulant/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbq            *database.Queries
	jwtSecret      string
	polkaKey       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	err := godotenv.Load() // Load environment variables from .env file

	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}

	dbQueries := database.New(db)

	jwtSecret := os.Getenv("JWT_SECRET")

	polkaKey := os.Getenv("POLKA_KEY")

	apiCfg := apiConfig{dbq: dbQueries, jwtSecret: jwtSecret, polkaKey: polkaKey}

	mux := http.NewServeMux()

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.Handle("/app/assets", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("GET /api/healthz", apiCfg.handleHealthz)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetAllChirps)

	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirpByID)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handleDeleteChirpByID)

	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)

	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)

	mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handleUgradeUserToChirpyRed)

	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)

	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshRefreshToken)

	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Server started on port 8080 baby!!!!!!")

	err = server.ListenAndServe()

	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
