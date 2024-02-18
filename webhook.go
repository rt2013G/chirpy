package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type webhookEvent struct {
	Event string         `json:"event"`
	Data  map[string]int `json:"data"`
}

const (
	USER_UPGRADED_EVENT = "user.upgraded"
)

func (apiCfg *apiConfig) userUpgrade(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := webhookEvent{}
	err := decoder.Decode(&params)
	if err != nil {
		ServerErrorResponse(w)
		return
	}

	requestKey := strings.TrimPrefix(r.Header.Get("authorization"), "ApiKey ")
	if len(requestKey) <= 0 {
		w.WriteHeader(401)
		return
	}

	polkaKey := []byte(os.Getenv("POLKA_KEY"))
	if requestKey != string(polkaKey) {
		w.WriteHeader(401)
		return
	}

	if params.Event != USER_UPGRADED_EVENT {
		w.WriteHeader(200)
		return
	}
	usrId := params.Data["user_id"]
	dbData := apiCfg.database.loadDB()
	user, ok := dbData.Users[usrId]
	if !ok {
		w.WriteHeader(404)
		return
	}
	user.IsChirpyRed = true
	dbData.Users[usrId] = user
	apiCfg.database.writeDB(dbData)
	w.WriteHeader(200)
}
