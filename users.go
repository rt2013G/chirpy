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
		ServerErrorResponse(w)
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
		ServerErrorResponse(w)
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

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * TOKEN_TTL_HOURS)),
		Subject:   fmt.Sprint(usrToSearch.Id),
	})
	token, err := claims.SignedString(apiCfg.jwtSecret)
	if err != nil {
		ServerErrorResponse(w)
		return
	}

	claims = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * REFRESH_TOKEN_TTL_HOURS)),
		Subject:   fmt.Sprint(usrToSearch.Id),
	})
	refresh_token, err := claims.SignedString(apiCfg.jwtSecret)
	if err != nil {
		ServerErrorResponse(w)
		return
	}

	resp := userResponseJWT{
		Id:           usrToSearch.Id,
		Email:        usrToSearch.Email,
		Token:        token,
		RefreshToken: refresh_token,
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
		ServerErrorResponse(w)
		return
	}

	auth := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(auth, &claims, func(token *jwt.Token) (interface{}, error) { return []byte(apiCfg.jwtSecret), nil })
	if err != nil {
		w.WriteHeader(401)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil || issuer == "chirpy-refresh" {
		w.WriteHeader(401)
		return
	}

	usrId, err := token.Claims.GetSubject()
	if err != nil {
		ServerErrorResponse(w)
		return
	}
	id, err := strconv.Atoi(usrId)
	if err != nil {
		ServerErrorResponse(w)
		return
	}
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(params.Password), 10)
	responseUserResource := apiCfg.database.UpdateUser(id, params.Email, passwordHash)

	data, _ := json.Marshal(responseUserResource)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (apiCfg *apiConfig) refreshUserToken(w http.ResponseWriter, r *http.Request) {
	receivedToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(receivedToken, &claims, func(token *jwt.Token) (interface{}, error) { return []byte(apiCfg.jwtSecret), nil })
	if err != nil {
		w.WriteHeader(401)
		return
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil || issuer != "chirpy-refresh" || apiCfg.database.isRevoked(receivedToken) {
		w.WriteHeader(401)
		return
	}

	usrId, err := token.Claims.GetSubject()
	if err != nil {
		ServerErrorResponse(w)
		return
	}
	id, err := strconv.Atoi(usrId)
	if err != nil {
		ServerErrorResponse(w)
		return
	}

	accessClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * TOKEN_TTL_HOURS)),
		Subject:   fmt.Sprint(id),
	})
	newAccessToken, err := accessClaims.SignedString(apiCfg.jwtSecret)
	if err != nil {
		ServerErrorResponse(w)
		return
	}

	accessToken := tokenResponse{
		Token: newAccessToken,
	}

	data, _ := json.Marshal(accessToken)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (apiCfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	receivedToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	apiCfg.database.revoke(receivedToken)
	w.WriteHeader(200)
}
