package main

import (
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










