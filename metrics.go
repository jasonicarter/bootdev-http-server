package main

import (
	"fmt"
	"log"
	"net/http"
)

func (cfg *apiConfig) getMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf(
		`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
		`,
		cfg.fileserverHits.Load())
	w.Write([]byte(msg))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
	}

	// Reset users
	err := cfg.dbQueries.ResetUsers(req.Context())
	if err != nil {
		// TODO: respond with error
		log.Printf("Error truncating Users table: %v", err)
	}
	err = cfg.dbQueries.ResetChirps(req.Context())
	if err != nil {
		log.Printf("Error truncating Chirps table: %v", err)
	}

	log.Printf("Tables truncated")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)

}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
