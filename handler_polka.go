package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
)

func (cfg *apiConfig) handlerPolka(w http.ResponseWriter, r *http.Request) {
	type dataStruct struct {
		UserID uuid.UUID `json:"user_id"`
	}
	type inVals struct {
		Event string `json:"event"`
		Data dataStruct `json:"data"`
	}

	inputs := inVals{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&inputs)
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
