package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) handlerUserUpdate(
	w http.ResponseWriter,
	r *http.Request,
) {
	type inVals struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	// * DRY 
	accessTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not find JWT")
		return
	}
	userID, err := auth.ValidateJWT(accessTokenString, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not validate JWT")
		return
	}
	// DRY *

	inputs := inVals{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&inputs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode inputs")
		return
	}

	hashedPassword, err := auth.HashPassword(inputs.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	err = cfg.db.SetPassword(r.Context(), database.SetPasswordParams{
		HashedPassword: hashedPassword,
		ID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not set password")
		return
	}
	err = cfg.db.SetEmail(r.Context(), database.SetEmailParams{
		Email: inputs.Email,
		ID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not set email")
		return
	}
	
	type outVals struct {
		User
	}
	usr, err := cfg.db.GetUserByEmailAddress(r.Context(), inputs.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get user")
		return
	}
	respondWithJSON(w, http.StatusOK, outVals{
		User: User{
			ID: usr.ID,
			IsChirpyRed: usr.IsChirpyRed,
			CreatedAt: usr.CreatedAt,
			UpdatedAt: usr.UpdatedAt,
			Email: usr.Email,
		},
	})
}



