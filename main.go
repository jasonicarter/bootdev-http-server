package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jasonicarter/bootdev-http-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

var bannedWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

func main() {

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	env := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		os.Exit(1)
	}

	apiCfg := apiConfig{
		// fileserverHits: // The zero value is zero
		dbQueries: database.New(db),
		platform:  env,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/",
		apiCfg.middlewareMetricsInc(
			http.StripPrefix("/app", http.FileServer(http.Dir("."))),
		),
	)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)

	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerAddChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpByID)
	mux.HandleFunc("POST /api/users", apiCfg.handlerAddUsers)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start things up and exist if it fails
	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Error starting up the server: %v", err)
		os.Exit(1)
	}

}

func replaceProfanity(msg string, bannedWords map[string]bool) string {

	words := strings.Fields(msg)
	for _, word := range words {
		// check if word in msg is a key in bannedWords and return value with is boolean
		if bannedWords[strings.ToLower(word)] {
			msg = strings.ReplaceAll(msg, word, "****")
		}
	}

	return msg
}

func respondWithError(w http.ResponseWriter, httpStatusCode int, msg string) {

	errorResponse := struct {
		Error string `json:"error"`
	}{Error: msg}
	respondWithJSON(w, httpStatusCode, errorResponse)

}

func respondWithJSON(w http.ResponseWriter, httpStatusCode int, payload any) {

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}

	w.WriteHeader(httpStatusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))

}

func handlerHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerAddChirp(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	// get json body into struct
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	// handle error parsing parameters
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// handle chirpy - validate length and respond
	// len() returns bytes not characters
	chirpLength := len([]rune(params.Body))

	// TODO: Consider moving this out into it's own func
	if chirpLength > 140 {
		errorResponse := struct {
			Error string `json:"error"`
		}{Error: "Chirp is too long"}

		respondWithJSON(w, http.StatusOK, errorResponse)
		return
	}

	if chirpLength <= 140 {

		chirpCleaned := replaceProfanity(params.Body, bannedWords)
		user_id, err := uuid.Parse(params.UserID)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		}
		newChirp := database.CreateChirpParams{
			Body:   chirpCleaned,
			UserID: user_id,
		}

		// Save chirp in database
		createdChirp, err := cfg.dbQueries.CreateChirp(req.Context(), newChirp)
		if err != nil {
			log.Printf("Error creating new chirp: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		// Respond
		type JSONResponse struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserID    uuid.UUID `json:"user_id"`
		}

		reqResponse := JSONResponse{
			ID:        createdChirp.ID,
			CreatedAt: createdChirp.CreatedAt,
			UpdatedAt: createdChirp.UpdatedAt,
			Body:      createdChirp.Body,
			UserID:    createdChirp.UserID,
		}
		respondWithJSON(w, http.StatusCreated, reqResponse)
		return
	}

}

func (cfg *apiConfig) handlerAddUsers(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// create user
	user, err := cfg.dbQueries.CreateUser(req.Context(), params.Email)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}

	// User struct allows for better control on the keys which database.user defaults
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	userCreated := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, userCreated)

}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, req *http.Request) {

	// Get chirps
	chirps, err := cfg.dbQueries.AllChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// Respond
	type JSONResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	var allChirps = []JSONResponse{}
	for _, c := range chirps {
		//add c variables into struct variable
		//append c to struct variable
		chirp := JSONResponse{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		}
		allChirps = append(allChirps, chirp)
	}

	respondWithJSON(w, http.StatusOK, allChirps)

}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, req *http.Request) {

	chirpId := req.PathValue("chirpID")
	if len(chirpId) == 0 {
		//TODO: respond with error
		return
	}
	log.Printf("%v", chirpId)

	id, err := uuid.Parse(chirpId)
	if err != nil {
		//TODO: respond with error
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(req.Context(), id)
	if err != nil {
		log.Printf("%v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// Respond
	type JSONResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	chirpByID := JSONResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.ID,
	}

	respondWithJSON(w, http.StatusOK, chirpByID)

}
