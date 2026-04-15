package server

import (
	"encoding/json"
	"net/http"
)

func nyx_AgentHandler(w http.ResponseWriter, r *http.Request) {

}

func nyx_AgentResolveHandler(w http.ResponseWriter, r *http.Request) {

}

type UserLogin struct {
	Username string `json:"user"`
	Password string `json:"passwd"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var log UserLogin
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, "invalid json", http.StatusInternalServerError)
		return
	}

	if !CheckLogin(log.Username, log.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid login"})
		return
	}

	token, err := CraftJWT(log.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "could not create token"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
