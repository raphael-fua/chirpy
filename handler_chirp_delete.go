package main

import (
	"chirpy/internal/auth"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
	chirpIDSTring := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDSTring)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
		return
	}
	
	accessTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "not authorized to delete")
		return
	}

	userID, err := auth.ValidateJWT(accessTokenString, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "not authorized to delete")
		return
	}

	chirp, err := cfg.db.GetOneChirp(r.Context(), chirpID)
    if err != nil {
		respondWithError(w, http.StatusNotFound, "could not get chirp")
		return
	}


	if userID != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "not authorized to delete")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "could not delete chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


