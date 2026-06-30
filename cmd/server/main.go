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
	mux.HandleFunc("GET /auth/strava", authStravaHandler)
	mux.HandleFunc("GET /auth/strava/callback", authStravaCallbackHandler)
	mux.HandleFunc("GET /athlete", athleteHandler)
	mux.HandleFunc("GET /api/integrations/strava", stravaIntegrationHandler)
	mux.HandleFunc("GET /api/activities", activitiesLocalHandler)
	mux.HandleFunc("POST /api/activities/sync", activitiesSyncHandler)
	mux.HandleFunc("GET /api/training-goal", trainingGoalGetHandler)
	mux.HandleFunc("PUT /api/training-goal", trainingGoalPutHandler)
	mux.HandleFunc("DELETE /api/training-goal", trainingGoalDeleteHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
