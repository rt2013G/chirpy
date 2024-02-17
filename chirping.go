package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (apiCfg *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := chirpParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		responseBody := errorResponseBody{
			Error: "Something went wrong",
		}
		data, _ := json.Marshal(responseBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(data)
		return
	}

	if len(params.Body) > 140 {
		responseBody := errorResponseBody{
			Error: "Chirp is too long",
		}
		data, _ := json.Marshal(responseBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	responseChirpResource := apiCfg.database.CreateChirp(cleanBody(params.Body))

	data, _ := json.Marshal(responseChirpResource)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (apiCfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "chirpID")
	var data []byte
	if len(param) > 0 {
		id, err := strconv.Atoi(param)
		if err != nil {
			ServerErrorResponse(w)
			return
		}
		chirp, ok := apiCfg.database.GetChirpWithId(id)
		if !ok {
			responseBody := errorResponseBody{
				Error: "Chirp doesn't exist",
			}
			data, _ := json.Marshal(responseBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(404)
			w.Write(data)
			return
		}
		data, _ = json.Marshal(chirp)
	} else {
		chirps, _ := apiCfg.database.GetChirps()
		data, _ = json.Marshal(chirps)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func cleanBody(body string) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		badWords := []string{"kerfuffle", "sharbert", "fornax"}
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}
