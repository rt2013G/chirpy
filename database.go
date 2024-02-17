package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps        map[int]fullChirpResource `json:"chirps"`
	Users         map[int]user              `json:"users"`
	RevokedTokens map[string]time.Time      `json:"revoked_tokens"`
}

type errorResponseBody struct {
	Error string `json:"error"`
}

type chirpParameters struct {
	Body string `json:"body"`
}

type fullChirpResource struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type user struct {
	Email        string `json:"email"`
	PasswordHash []byte `json:"password_hash"`
	Id           int    `json:"id"`
}

type userParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userResponse struct {
	Email string `json:"email"`
	Id    int    `json:"id"`
}

type userResponseJWT struct {
	Email        string `json:"email"`
	Id           int    `json:"id"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

const (
	TOKEN_TTL_HOURS         = 1
	REFRESH_TOKEN_TTL_HOURS = 1440
)

func NewDB(path string) (*DB, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		dbStructure := DBStructure{
			Chirps:        map[int]fullChirpResource{},
			Users:         map[int]user{},
			RevokedTokens: map[string]time.Time{},
		}
		jsonDb, _ := json.Marshal(dbStructure)
		err := os.WriteFile(path, jsonDb, 0666)
		if err != nil {
			return nil, nil
		}
	}

	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	db.loadDB()
	return &db, nil
}

func (db *DB) loadDB() DBStructure {
	db.mux.RLock()
	data, err := os.ReadFile(db.path)
	if err != nil {
		log.Fatal("error while reading db file")
	}
	db.mux.RUnlock()
	chirpData := DBStructure{}
	err = json.Unmarshal(data, &chirpData)
	if err != nil {
		log.Fatal("error while reading db file")
	}
	return chirpData
}

func (db *DB) GetChirps() ([]fullChirpResource, error) {
	dbStructure := db.loadDB()
	chirpsList := make([]fullChirpResource, 0)
	for _, chirp := range dbStructure.Chirps {
		chirpsList = append(chirpsList, chirp)
	}

	return chirpsList, nil
}

func (db *DB) GetChirpWithId(id int) (fullChirpResource, bool) {
	chirps := db.loadDB()
	chirp, ok := chirps.Chirps[id]
	if !ok {
		return fullChirpResource{}, false
	}

	return chirp, true
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	db.mux.Lock()
	err = os.WriteFile(db.path, data, 0666)
	db.mux.Unlock()
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateChirp(body string) fullChirpResource {
	chirpData := db.loadDB()
	newChirp := fullChirpResource{
		Body: body,
		Id:   len(chirpData.Chirps) + 1,
	}
	chirpData.Chirps[newChirp.Id] = newChirp
	db.writeDB(chirpData)
	return newChirp
}

func (db *DB) CreateUser(email string, passwordHash []byte) userResponse {
	dbData := db.loadDB()
	newUser := user{
		Email:        email,
		PasswordHash: passwordHash,
		Id:           len(dbData.Users) + 1,
	}
	dbData.Users[newUser.Id] = newUser
	db.writeDB(dbData)

	userResp := userResponse{
		Email: email,
		Id:    newUser.Id,
	}
	return userResp
}

func (db *DB) UpdateUser(id int, email string, passwordHash []byte) userResponse {
	dbData := db.loadDB()
	user := dbData.Users[id]
	user.Email = email
	user.PasswordHash = passwordHash
	dbData.Users[id] = user
	db.writeDB(dbData)

	userResp := userResponse{
		Email: email,
		Id:    id,
	}
	return userResp
}

func ServerErrorResponse(w http.ResponseWriter) {
	responseBody := errorResponseBody{
		Error: "Something went wrong",
	}
	data, _ := json.Marshal(responseBody)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	w.Write(data)
}

func (db *DB) isRevoked(token string) bool {
	dbData := db.loadDB()
	_, ok := dbData.RevokedTokens[token]
	return ok
}

func (db *DB) revoke(token string) {
	dbData := db.loadDB()
	dbData.RevokedTokens[token] = time.Now().UTC()
	db.writeDB(dbData)
}
