package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestActivitiesHandler_FetchesStoresAndReturnsActivities(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/athlete/activities" {
			t.Fatalf("Expected /athlete/activities path, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "1" {
			t.Fatalf("Expected page 1, got %q", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("per_page") != "200" {
			t.Fatalf("Expected per_page 200, got %q", r.URL.Query().Get("per_page"))
		}
		if r.Header.Get("Authorization") != "Bearer valid-access-token" {
			t.Fatalf("Expected bearer token header, got %q", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"id": 123,
				"type": "Run",
				"sport_type": "Run",
				"name": "Morning Run",
				"start_date": "2026-05-15T06:30:00Z",
				"elapsed_time": 3600,
				"moving_time": 3500,
				"distance": 10000.5,
				"total_elevation_gain": 50.2,
				"average_heartrate": 145.5,
				"max_heartrate": 180,
				"calories": 650
			}
		]`))
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	originalProviderTokenStore := providerTokenStore
	originalProviderRawObjectStore := providerRawObjectStore
	originalProviderActivityStore := providerActivityStore
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
		providerTokenStore = originalProviderTokenStore
		providerRawObjectStore = originalProviderRawObjectStore
		providerActivityStore = originalProviderActivityStore
	})

	stravaAPIBaseURL = stravaServer.URL
	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}
	rawStore := &fakeRawObjectStore{}
	activityStore := &fakeActivityStore{}
	providerRawObjectStore = rawStore
	providerActivityStore = activityStore

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200 OK, got %d: %s", resp.StatusCode, string(body))
	}

	if rawStore.object.ObjectKey == "" {
		t.Fatal("Expected raw Strava response to be stored")
	}

	if activityStore.saveCount != 1 {
		t.Fatalf("Expected activities to be saved once, got %d saves", activityStore.saveCount)
	}

	if len(activityStore.activities) != 1 {
		t.Fatalf("Expected one stored activity, got %d", len(activityStore.activities))
	}

	stored := activityStore.activities[0]
	if stored.ProviderID != "123" {
		t.Fatalf("Expected provider id 123, got %q", stored.ProviderID)
	}
	if stored.ActivityType != "run" {
		t.Fatalf("Expected normalized activity type run, got %q", stored.ActivityType)
	}
	if stored.RawObjectKey != rawStore.object.ObjectKey {
		t.Fatalf("Expected raw object key %q, got %q", rawStore.object.ObjectKey, stored.RawObjectKey)
	}

	var response []Activity
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if len(response) != 1 {
		t.Fatalf("Expected one response activity, got %d", len(response))
	}
}

func TestActivitiesHandler_ReturnsErrorWhenTokenIsMissing(t *testing.T) {
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		providerTokenStore = originalProviderTokenStore
	})

	providerTokenStore = &fakeTokenStore{err: errors.New("token not found")}

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}

func TestActivitiesHandler_ReturnsTooManyRequestsWhenStravaRateLimitIsExceeded(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
		providerTokenStore = originalProviderTokenStore
	})

	stravaAPIBaseURL = stravaServer.URL
	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429 Too Many Requests, got %d", resp.StatusCode)
	}
}

func TestActivitiesHandler_ReturnsBadGatewayWhenStravaJSONIsInvalid(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{`))
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
		providerTokenStore = originalProviderTokenStore
	})

	stravaAPIBaseURL = stravaServer.URL
	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("Expected status 502 Bad Gateway, got %d", resp.StatusCode)
	}
}

func TestActivitiesHandler_ReturnsErrorWhenRawStoreFails(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	originalProviderTokenStore := providerTokenStore
	originalProviderRawObjectStore := providerRawObjectStore
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
		providerTokenStore = originalProviderTokenStore
		providerRawObjectStore = originalProviderRawObjectStore
	})

	stravaAPIBaseURL = stravaServer.URL
	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}
	providerRawObjectStore = &fakeRawObjectStore{err: errors.New("raw store failed")}

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}

func TestActivitiesHandler_ReturnsBadGatewayWhenMappingFails(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"id": 123,
				"type": "Run",
				"start_date": "2026-05-15T06:30:00Z"
			}
		]`))
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	originalProviderTokenStore := providerTokenStore
	originalProviderRawObjectStore := providerRawObjectStore
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
		providerTokenStore = originalProviderTokenStore
		providerRawObjectStore = originalProviderRawObjectStore
	})

	stravaAPIBaseURL = stravaServer.URL
	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}
	providerRawObjectStore = &fakeRawObjectStore{}

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("Expected status 502 Bad Gateway, got %d", resp.StatusCode)
	}
}

func TestActivitiesHandler_ReturnsErrorWhenActivityStoreSaveFails(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"id": 123,
				"type": "Run",
				"name": "Morning Run",
				"start_date": "2026-05-15T06:30:00Z"
			}
		]`))
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	originalProviderTokenStore := providerTokenStore
	originalProviderRawObjectStore := providerRawObjectStore
	originalProviderActivityStore := providerActivityStore
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
		providerTokenStore = originalProviderTokenStore
		providerRawObjectStore = originalProviderRawObjectStore
		providerActivityStore = originalProviderActivityStore
	})

	stravaAPIBaseURL = stravaServer.URL
	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}
	providerRawObjectStore = &fakeRawObjectStore{}
	providerActivityStore = &fakeActivityStore{err: errors.New("activity store failed")}

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}

func TestActivitiesHandler_ReturnsErrorWhenActivityStoreListFails(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"id": 123,
				"type": "Run",
				"name": "Morning Run",
				"start_date": "2026-05-15T06:30:00Z"
			}
		]`))
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	originalProviderTokenStore := providerTokenStore
	originalProviderRawObjectStore := providerRawObjectStore
	originalProviderActivityStore := providerActivityStore
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
		providerTokenStore = originalProviderTokenStore
		providerRawObjectStore = originalProviderRawObjectStore
		providerActivityStore = originalProviderActivityStore
	})

	stravaAPIBaseURL = stravaServer.URL
	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}
	providerRawObjectStore = &fakeRawObjectStore{}
	providerActivityStore = &fakeActivityStore{listErr: errors.New("activity list failed")}

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}

func TestActivitiesHandler_DoesNotExposeTokenWhenStravaFetchFails(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
		providerTokenStore = originalProviderTokenStore
	})

	stravaAPIBaseURL = stravaServer.URL
	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "secret-access-token",
			RefreshToken: "secret-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}

	req := httptest.NewRequest("GET", "/api/activities", nil)
	w := httptest.NewRecorder()

	activitiesSyncHandler(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response body: %v", err)
	}

	if strings.Contains(string(body), "secret-access-token") || strings.Contains(string(body), "secret-refresh-token") {
		t.Fatalf("Expected response body not to expose token values, got %q", string(body))
	}
}
