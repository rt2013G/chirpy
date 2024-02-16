package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.HandleFunc("/healthz", handlerReadiness)

	corsMux := middlewareCors(serveMux)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	log.Fatal(server.ListenAndServe())
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
