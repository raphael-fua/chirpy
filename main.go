package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


var badWords = map[string]bool {
	"kerfuffle": true,
	"sharbert": true,
	"fornax": true,
}


type User struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string `json:"email"`

}


type apiConfig struct {
	fileserverHits atomic.Int32
	queries *database.Queries
	platform string
	jwtsecret string
}


func main() {
	godotenv.Load()
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil { log.Fatal(err) }
	dbQueries := database.New(db)

	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		queries: dbQueries,
		platform:os.Getenv("PLATFORM"),
		jwtsecret: os.Getenv("JWT_SECRET"),
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

		
		

	mux.HandleFunc(
		"GET /api/healthz",
		func(w http.ResponseWriter, r *http.Request){
 			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(http.StatusText(http.StatusOK)))
		})

	mux.HandleFunc(
		"GET /admin/metrics", 
		func(w http.ResponseWriter, r *http.Request){
 			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", apiCfg.fileserverHits.Load())))
		})

	
	mux.HandleFunc(
		"POST /api/users",
		func(w http.ResponseWriter, r *http.Request) {

            type parameters struct {
				Email string `json:"email"`
				Password string `json:"password"`
			}

			decoder := json.NewDecoder(r.Body)
			params := parameters{}
			err := decoder.Decode(&params)
			if err != nil {
				respondWithError(
					w, http.StatusBadRequest, "Error decoding (client request error)")
				return
			}

			hashedPassword, err := auth.HashPassword(params.Password)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError,
					"problem hashing password")
			}

			userWild, err := dbQueries.CreateUser(r.Context(), params.Email)
            if err != nil {
				respondWithError(w, http.StatusInternalServerError,
					"Error creating user (server failure)")
				return
			}

			err = dbQueries.SetPassword(r.Context(), database.SetPasswordParams{
				HashedPassword: hashedPassword,
				ID: userWild.ID,
			})

			if err != nil {
				respondWithError(w, http.StatusInternalServerError,
					"problem saving hashed password to database")
			}

			userTame := User{
				ID: userWild.ID,
				CreatedAt: userWild.CreatedAt,
				UpdatedAt: userWild.UpdatedAt,
				Email: userWild.Email,
			}
				respondWithJSON(w, http.StatusCreated, userTame)
		})




	mux.HandleFunc(
		"GET /api/chirps/{chirpID}",
		func(w http.ResponseWriter, r *http.Request) {
			path := r.PathValue("chirpID")
			id, err := uuid.Parse(path)
			if err != nil {
				respondWithError(
					w, http.StatusBadRequest, "Error: entered id not valid")
				return
			}

			chirp, err := dbQueries.GetOneChirp(r.Context(), id)
			if err != nil {
				respondWithError(
					w, http.StatusNotFound, "Error: no chirp with that id found")
				return
			}
			type outChirp struct {
				ID uuid.UUID `json:"id"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
				Body string `json:"body"`
				UserID uuid.UUID `json:"user_id"`
			}
			respondWithJSON(w, http.StatusOK, outChirp{
				ID: chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body: chirp.Body,
				UserID: chirp.UserID,
			})
		})


	mux.HandleFunc(
		"GET /api/chirps",
		func(w http.ResponseWriter, r *http.Request) {
			chirps, err := dbQueries.GetAllChirps(r.Context())
			if err != nil {
				respondWithError(
					w, http.StatusInternalServerError, "Error on `GetAllChirps` call")
				return
			}
			type outChirp struct {
				ID uuid.UUID `json:"id"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
				Body string `json:"body"`
				UserID uuid.UUID `json:"user_id"`
			}
			outchirps := []outChirp{}
			for _, chirp := range chirps {
				outchirps = append(outchirps, outChirp{
					ID: chirp.ID,
					CreatedAt: chirp.CreatedAt,
					UpdatedAt: chirp.UpdatedAt,
					Body: chirp.Body,
					UserID: chirp.UserID,
				})
			}
			respondWithJSON(w, http.StatusOK, outchirps)
		})


	mux.HandleFunc(
		"POST /api/login",
		func(w http.ResponseWriter, r *http.Request) {
			type inVals struct {
				Email string `json:"email"`
				Password string `json:"password"`
				ExpiresInSeconds *int `json:"expires_in_seconds"`
			}
			type outVals struct {
				ID uuid.UUID `json:"id"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
				Email string `json:"email"`
				TokenString string `json:"token"`
			}
			returnExpiration := 3600
			decoder := json.NewDecoder(r.Body)
			invals := inVals{}
			err := decoder.Decode(&invals)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "error decoding request body")
				return
			}
			if invals.ExpiresInSeconds != nil {
				returnExpiration = min(returnExpiration, *invals.ExpiresInSeconds)
			}
			usr, err := dbQueries.GetUserByEmailAddress(r.Context(), invals.Email)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "error finding email")
				return
			}
			match, err := auth.CheckPasswordHash(invals.Password, usr.HashedPassword)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, 
					"error checking password hash")
				return
			}
			if !match {
				respondWithError(w, http.StatusUnauthorized, "error wrong password")
				return
			}
			tokenString, err := auth.MakeJWT(
				usr.ID, apiCfg.jwtsecret, time.Duration(returnExpiration) * time.Second)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError,
					"error creating token string")
				return
			}
			respondWithJSON(w, http.StatusOK, outVals{
				ID: usr.ID,
				CreatedAt: usr.CreatedAt,
				UpdatedAt: usr.UpdatedAt,
				Email: usr.Email,
				TokenString: tokenString,
			})
		})


	mux.HandleFunc(
		"POST /api/chirps",
		func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := auth.GetBearerToken(r.Header)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "error getting token string")
				return
			}
			userID, err := auth.ValidateJWT(tokenString, apiCfg.jwtsecret)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "error invalid JWT")
				return
			}
			type inVals struct {
				Body string `json:"body"`
				// UserID uuid.UUID `json:"user_id"`
			}
			type outVals struct {
				ID uuid.UUID `json:"id"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
				Body string `json:"body"`
				UserID uuid.UUID `json:"user_id"`
			}
			decoder := json.NewDecoder(r.Body)
			in := inVals{}
			err = decoder.Decode(&in)
			if err != nil {
				respondWithError(w, 500, "Error decoding parameter")
				return
			}
			if len(in.Body) > 140 {
				respondWithError(w, 400, "Chirp is too long")
				return
			}
			words := strings.Split(in.Body, " ")
			for i, word := range words {
				if badWords[strings.ToLower(word)] {
					words[i] = "****"
				}
			}
			chirpWild, err := dbQueries.CreateChirp(
				r.Context(), database.CreateChirpParams{
					Body: strings.Join(words, " "),
					UserID: userID,
				})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError,
					"Error creating chirp (server failure)")
				return
			}
			respondWithJSON(w, 201, outVals{
				ID: chirpWild.ID,
				CreatedAt: chirpWild.CreatedAt,
				UpdatedAt: chirpWild.UpdatedAt,
				Body: chirpWild.Body,
				UserID: chirpWild.UserID,
			})
		})

	mux.HandleFunc(
		"POST /admin/reset", 
		func(w http.ResponseWriter, r *http.Request){
			if apiCfg.platform != "dev" {
				respondWithError(
					w, http.StatusForbidden, "reset can only be used in dev mode")
				return
			}
			apiCfg.fileserverHits.Store(0)
			dbQueries.ResetUsers(r.Context())
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hits reset to 0\n"))
		})


	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}


func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}


func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorReturnVal struct {
		Message string `json:"error"`
	}
	respBody := errorReturnVal {
        Message: msg,
	}
	respondWithJSON(w, code, respBody)
}












