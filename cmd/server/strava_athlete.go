package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Athlete struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Sex       string `json:"sex"`
	Premium   bool   `json:"premium"`
}

func fetchStravaAthlete(accessToken string) (*Athlete, error) {
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	req, err := http.NewRequest(
		http.MethodGet,
		stravaAPIBaseURL+"/athlete",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating Strava athlete request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := stravaHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Strava athlete endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava athlete request failed with status %d", resp.StatusCode)
	}

	var athlete Athlete
	err = json.NewDecoder(resp.Body).Decode(&athlete)
	if err != nil {
		return nil, fmt.Errorf("decoding Strava athlete response: %w", err)
	}
	return &athlete, nil
}

func athleteHandler(w http.ResponseWriter, r *http.Request) {
	cfg := loadAppConfig()
	token, err := getValidStravaToken(r.Context(), cfg.AppUserID)
	if err != nil {
		http.Error(w, "Strava token is not configured", http.StatusInternalServerError)
		return
	}

	athlete, err := fetchStravaAthlete(token.AccessToken)
	if err != nil {
		http.Error(w, "failed to fetch Strava athlete", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(athlete)
}
