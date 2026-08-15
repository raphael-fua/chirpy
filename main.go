package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}



func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
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

	// mux.HandleFunc(
	// 	"GET /api/metrics", 
	// 	func(w http.ResponseWriter, r *http.Request){
	// 			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	// 		w.WriteHeader(http.StatusOK)
	// 		w.Write([]byte(fmt.Sprintf("Hits: %d\n", apiCfg.fileserverHits.Load())))
	// 	})
	mux.HandleFunc(
		"GET /admin/metrics", 
		func(w http.ResponseWriter, r *http.Request){
 			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", apiCfg.fileserverHits.Load())))
		})




	mux.HandleFunc(
	    "POST /api/validate_chirp",
		func(w http.ResponseWriter, r *http.Request) {
            type parameter struct {
                Body string `json:"body"`
			}
			type validity struct {
				Valid bool `json:"valid"`
			}

	        decoder := json.NewDecoder(r.Body)
			param := parameter{}
		    err := decoder.Decode(&param)
		    if err != nil {
				respondWithError(w, 500, "Error decoding parameter")
				return
			}

			if len(param.Body) > 140 {
				respondWithError(w, 400, "Chirp is too long")
				return
			}

			respondWithJSON(w, 200, validity{Valid: true})
		})

	mux.HandleFunc(
		"POST /admin/reset", 
		func(w http.ResponseWriter, r *http.Request){
			apiCfg.fileserverHits.Store(0)
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









