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

	apiCfg := apiConfig{dbq: dbQueries}

	mux := http.NewServeMux()

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.Handle("/app/assets", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("GET /api/healthz", handleHealthz)

	mux.HandleFunc("GET /admin/metrics", handleMetrics(&apiCfg))

	mux.HandleFunc("POST /admin/reset", handleReset(&apiCfg))

	mux.HandleFunc("POST /api/validate_chirp", handleValidateBodyLength)

	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)

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
