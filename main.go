package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var bannedWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

func main() {

	apiCfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/",
		apiCfg.middlewareMetricsInc(
			http.StripPrefix("/app", http.FileServer(http.Dir("."))),
		),
	)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()

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

func healthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func validateChirp(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		Body string `json:"body"`
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
	if chirpLength <= 140 {

		chirpCleaned := replaceProfanity(params.Body, bannedWords)
		validResponse := struct {
			CleanBody string `json:"cleaned_body"`
		}{CleanBody: chirpCleaned}

		respondWithJSON(w, http.StatusOK, validResponse)
		return
	}

	if chirpLength > 140 {
		errorResponse := struct {
			Error string `json:"error"`
		}{Error: "Chirp is too long"}

		respondWithJSON(w, http.StatusOK, errorResponse)
		return
	}

}
