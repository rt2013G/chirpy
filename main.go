package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	database       *DB
	fileserverHits int
	jwtSecret      []byte
}

func main() {
	godotenv.Load()
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if *debug {
		if err := os.Remove("database.json"); err != nil {
			log.Fatal(err)
		}
	}

	port := os.Getenv("PORT")
	dbPath := "database.json"

	db, err := NewDB(dbPath)
	if err != nil {
		log.Fatal("error while creating db")
	}

	apiCfg := apiConfig{
		database:       db,
		fileserverHits: 0,
		jwtSecret:      []byte(os.Getenv("JWT_SECRET")),
	}

	router := chi.NewRouter()

	fileserverHandlers := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	router.Get("/", apiCfg.handlerIndex)
	router.Handle("/app", fileserverHandlers)
	router.Handle("/app/*", fileserverHandlers)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)
	apiRouter.Post("/chirps", apiCfg.postChirp)
	apiRouter.Get("/chirps", apiCfg.getChirps)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.getChirps)
	apiRouter.Delete("/chirps/{chirpID}", apiCfg.deleteChirp)
	apiRouter.Post("/users", apiCfg.postUser)
	apiRouter.Put("/users", apiCfg.putUser)
	apiRouter.Post("/login", apiCfg.loginUser)
	apiRouter.Post("/refresh", apiCfg.refreshUserToken)
	apiRouter.Post("/revoke", apiCfg.revokeToken)
	apiRouter.Post("/polka/webhooks", apiCfg.userUpgrade)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.handlerMetrics)
	router.Mount("/admin", adminRouter)

	corsMux := middlewareCors(router)

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

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>

	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>`, cfg.fileserverHits)))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset"))
}

func (cfg *apiConfig) handlerIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>

	<body>
		<h1>Welcome to Chirpy </h1>
		<p> Hello from Docker! I'm a Go server. </p>
	</body>
	
	</html>`)))
}
