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
	r, _ := cfg.dbQueries.ResetUsers(req.Context())

	// TODO: err contains "no rows in result set"
	// if err != nil {
	// 	log.Printf("Error reseting users: %v", err)
	// 	respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	// }
	log.Printf("Users table truncates: %v", r)

	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)

}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
