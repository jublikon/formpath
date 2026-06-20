package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var providerRawObjectStore RawObjectStore = noopRawObjectStore{}
var providerActivityStore ActivityStore = noopActivityStore{}

func activitiesLocalHandler(w http.ResponseWriter, r *http.Request) {
	cfg := loadAppConfig()
	activities, err := providerActivityStore.ListActivities(r.Context(), cfg.AppUserID)
	if err != nil {
		http.Error(w, "failed to list activities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}

func activitiesSyncHandler(w http.ResponseWriter, r *http.Request) {
	cfg := loadAppConfig()
	token, err := getValidStravaToken(r.Context(), cfg.AppUserID)
	if err != nil {
		http.Error(w, "Strava token is not configured", http.StatusInternalServerError)
		return
	}

	fetchedAt := time.Now().UTC()
	rawBody, err := fetchStravaActivitiesPayload(token.AccessToken)
	if err != nil {
		var statusErr HTTPStatusError
		if errors.As(err, &statusErr) {
			switch statusErr.code {
			case http.StatusTooManyRequests:
				http.Error(w, "Strava API rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			case http.StatusUnauthorized:
				http.Error(w, "Strava token was rejected", http.StatusBadGateway)
				return
			}
		}
		http.Error(w, "failed to fetch Strava activities", http.StatusBadGateway)
		return
	}

	rawObjectKey := stravaActivitiesObjectKey(cfg.AppUserID, fetchedAt)
	err = providerRawObjectStore.SaveRawObject(r.Context(), RawObject{
		UserID:             cfg.AppUserID,
		Provider:           "strava",
		ProviderObjectType: "activity_list",
		ProviderObjectID:   rawObjectKey,
		ObjectKey:          rawObjectKey,
		ContentType:        "application/json",
		Body:               rawBody,
		FetchedAt:          fetchedAt,
	})
	if err != nil {
		http.Error(w, "failed to store raw Strava activities", http.StatusInternalServerError)
		return
	}

	loadedRawBody, err := providerRawObjectStore.GetRawObject(r.Context(), rawObjectKey)
	if err != nil {
		http.Error(w, "failed to load raw Strava activities", http.StatusInternalServerError)
		return
	}

	activities, err := transformStravaActivities(cfg.AppUserID, loadedRawBody, rawObjectKey)
	if err != nil {
		http.Error(w, "failed to transform Strava activities", http.StatusBadGateway)
		return
	}

	if err := providerActivityStore.SaveActivities(r.Context(), activities); err != nil {
		http.Error(w, "failed to store activities", http.StatusInternalServerError)
		return
	}

	storedActivities, err := providerActivityStore.ListActivities(r.Context(), cfg.AppUserID)
	if err != nil {
		http.Error(w, "failed to list activities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(storedActivities)
}
