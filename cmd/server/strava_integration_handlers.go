package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

type stravaIntegrationResponse struct {
	Provider  string `json:"provider"`
	Connected bool   `json:"connected"`
}

func stravaIntegrationHandler(w http.ResponseWriter, r *http.Request) {
	cfg := loadAppConfig()

	token, err := providerTokenStore.GetProviderToken(r.Context(), cfg.AppUserID, "strava")
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "failed to load Strava integration", http.StatusInternalServerError)
			return
		}
	}

	response := stravaIntegrationResponse{
		Provider:  "strava",
		Connected: err == nil && token.RefreshToken != "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
