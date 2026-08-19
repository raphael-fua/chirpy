package main

import (
	"chirpy/internal/auth"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolka(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not api key in header")
		return
	}
	if cfg.apiKey != apiKey {
		respondWithError(w, http.StatusUnauthorized, "wrong api key")
		return
	}

	type dataStruct struct {
		UserID uuid.UUID `json:"user_id"`
	}
	type inVals struct {
		Event string `json:"event"`
		Data dataStruct `json:"data"`
	}

	inputs := inVals{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&inputs)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not decode request body")
		return
	}

	if inputs.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.db.UpGradeUserToRed(r.Context(), inputs.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "user to upgrade not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
