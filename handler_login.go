package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type inVals struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}
	type outVals struct {
		User
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	inputs := inVals{}
	err := decoder.Decode(&inputs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode inputs")
		return
	}

	usr, err := cfg.db.GetUserByEmailAddress(r.Context(), inputs.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(inputs.Password, usr.HashedPassword)
	if err != nil || !match{
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	accessToken, err := auth.MakeJWT(usr.ID, cfg.jwtsecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create access token")
		return
	}

	refreshToken := auth.MakeRefreshToken()

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    usr.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not save refresh token")
		return
	}

	respondWithJSON(w, http.StatusOK, outVals{
		User: User{
			ID:        usr.ID,
			IsChirpyRed: usr.IsChirpyRed,
			CreatedAt: usr.CreatedAt,
			UpdatedAt: usr.UpdatedAt,
			Email:     usr.Email,
		},
		Token:         accessToken,
        RefreshToken:  refreshToken,
	})
}







