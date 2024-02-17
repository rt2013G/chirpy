package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) postUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := userParams{}
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
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(params.Password), 10)
	responseUserResource := db.CreateUser(params.Email, passwordHash)

	data, _ := json.Marshal(responseUserResource)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (db *DB) loginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := userParams{}
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
	dbStructure := db.loadDB()
	var usrToSearch user
	for _, val := range dbStructure.Users {
		if val.Email == params.Email {
			usrToSearch = val
			break
		}
	}
	err = bcrypt.CompareHashAndPassword(usrToSearch.PasswordHash, []byte(params.Password))
	if err != nil {
		w.WriteHeader(401)
		return
	}
	resp := userResponse{
		Id:    usrToSearch.Id,
		Email: usrToSearch.Email,
	}
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
