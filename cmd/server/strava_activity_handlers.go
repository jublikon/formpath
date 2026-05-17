package main

import (
	"encoding/json"
	"net/http"
	"time"
)

var providerRawObjectStore RawObjectStore = noopRawObjectStore{}
var providerActivityStore ActivityStore = noopActivityStore{}

func activitiesHandler(w http.ResponseWriter, r *http.Request) {
	cfg := loadAppConfig()
	token, err := getValidStravaToken(r.Context(), cfg.AppUserID)
	if err != nil {
		http.Error(w, "Strava token is not configured", http.StatusInternalServerError)
		return
	}

	fetchedAt := time.Now().UTC()
	stravaActivities, rawBody, err := fetchStravaActivities(token.AccessToken)
	if err != nil {
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

	activities, err := mapStravaActivities(cfg.AppUserID, stravaActivities, rawObjectKey)
	if err != nil {
		http.Error(w, "failed to map Strava activities", http.StatusBadGateway)
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
