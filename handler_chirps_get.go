package main

import (
	"github.com/google/uuid"
	"net/http"
	"sort"
)

func (cfg *apiConfig) handlerChirpGet(w http.ResponseWriter, r *http.Request) {
	chirpIDSTring := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDSTring)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
		return
	}

	dbChirp, err := cfg.db.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "could not get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID: dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		UserID: dbChirp.UserID,
		Body: dbChirp.Body,
	})
}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	headerAuthorIDString := r.URL.Query().Get("author_id")
	headerSortString := r.URL.Query().Get("sort")
	if headerSortString == "" {
		headerSortString = "asc"
	}

	dbChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not retrieve chirps")
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		if headerAuthorIDString != "" {
			headerAuthorID, err := uuid.Parse(headerAuthorIDString)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "could not parse author id")
				return
			}
			if dbChirp.UserID != headerAuthorID {
				continue
			}
		}
		chirps = append(chirps, Chirp{
			ID: dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID: dbChirp.UserID,
			Body: dbChirp.Body,
		})
	}

	if headerSortString == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	} else if headerSortString == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[j].CreatedAt.Before(chirps[i].CreatedAt)
		})
	} else {
		respondWithError(w, http.StatusBadRequest, "neither asc nor desc")
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)
}








