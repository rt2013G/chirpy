package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (apiCfg *apiConfig) postUser(w http.ResponseWriter, r *http.Request) {
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
	responseUserResource := apiCfg.database.CreateUser(params.Email, passwordHash)

	data, _ := json.Marshal(responseUserResource)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (apiCfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
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
	dbStructure := apiCfg.database.loadDB()
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
	if params.TTL == 0 {
		params.TTL = 86400
	}
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * params.TTL)),
		Subject:   fmt.Sprint(usrToSearch.Id),
	})
	token, err := claims.SignedString(apiCfg.jwtSecret)
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

	resp := userResponseJWT{
		Id:    usrToSearch.Id,
		Email: usrToSearch.Email,
		Token: token,
	}
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (apiCfg *apiConfig) putUser(w http.ResponseWriter, r *http.Request) {
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

	auth := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(auth, &claims, func(token *jwt.Token) (interface{}, error) { return []byte(apiCfg.jwtSecret), nil })
	if err != nil {
		w.WriteHeader(401)
		return
	}
	usrId, err := token.Claims.GetSubject()
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
	id, err := strconv.Atoi(usrId)
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
	responseUserResource := apiCfg.database.UpdateUser(id, params.Email, passwordHash)

	data, _ := json.Marshal(responseUserResource)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
