package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type DB struct {
	path   string
	mux    *sync.RWMutex
	nextId int
}

type DBStructure struct {
	Chirps map[int]fullChirpResource `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		dbStructure := DBStructure{
			Chirps: map[int]fullChirpResource{},
		}
		jsonDb, _ := json.Marshal(dbStructure)
		err := os.WriteFile(path, jsonDb, 0666)
		if err != nil {
			return nil, nil
		}
	}

	db := DB{
		path:   path,
		mux:    &sync.RWMutex{},
		nextId: 1,
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
	chirps := db.loadDB()
	chirpsList := make([]fullChirpResource, 0)
	for _, chirp := range chirps.Chirps {
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
	newChirp := fullChirpResource{
		Body: body,
		Id:   db.nextId,
	}
	db.nextId++
	chirpData := db.loadDB()
	chirpData.Chirps[newChirp.Id] = newChirp
	db.writeDB(chirpData)
	return newChirp
}
