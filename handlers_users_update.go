package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type updateResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	logCall(r)
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(cfg.jwt), nil },
	)
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if issuer == Refresh {
		respondWithError(w, http.StatusUnauthorized, "Issuer is Refresh.")
		return
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expiresAt, err := token.Claims.GetExpirationTime()
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if expiresAt.Before(time.Now().UTC()) {
		log.Printf(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	id, err := strconv.Atoi(userIDString)
	user, err := cfg.DB.UpdateUser(id, params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Update Failed")
	}
	respondWithJSON(w, http.StatusOK, updateResponse{ID: id, Email: user.Email})
}
