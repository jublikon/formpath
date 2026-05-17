package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := loadAppConfig()

	persistence, err := configurePersistence(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer persistence.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/auth/strava", authStravaHandler)
	mux.HandleFunc("/auth/strava/callback", authStravaCallbackHandler)
	mux.HandleFunc("/athlete", athleteHandler)
	mux.HandleFunc("/api/activities", activitiesLocalHandler)
	mux.HandleFunc("/api/activities/sync", activitiesSyncHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
