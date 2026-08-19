package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type User struct {
	ID           uuid.UUID `json:"id"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
}



func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type inVals struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}
	type outVals struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
    inputs := inVals{}
	err := decoder.Decode(&inputs)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding (client request error)")
		return
	}

	hashedPassword, err := auth.HashPassword(inputs.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "problem hashing password")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          inputs.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	respondWithJSON(w, http.StatusCreated, outVals{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}




