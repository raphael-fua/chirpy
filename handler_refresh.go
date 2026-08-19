package main

import (
	"chirpy/internal/auth"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type outVals struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not find refresh token")
		return
	}
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(),refreshToken)
	if err != nil {
		respondWithError(
			w,
			http.StatusUnauthorized,
			"could not get user for refresh token",
		)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtsecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not validate token")
		return
	}

	respondWithJSON(w, http.StatusOK, outVals{
		Token: accessToken,
	})
}

