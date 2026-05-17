package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchStravaAthlete(t *testing.T) {
	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/athlete" {
			t.Fatalf("Expected /athlete path, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer valid-access-token" {
			t.Fatalf("Expected bearer token header, got %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("Expected Accept application/json, got %q", r.Header.Get("Accept"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": 42,
			"username": "runner42",
			"firstname": "Ada",
			"lastname": "Runner",
			"city": "Berlin",
			"country": "Germany",
			"sex": "F",
			"premium": true
		}`))
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
	})
	stravaAPIBaseURL = stravaServer.URL

	athlete, err := fetchStravaAthlete("valid-access-token")
	if err != nil {
		t.Fatalf("Expected athlete, got error: %v", err)
	}

	if athlete.ID != 42 {
		t.Fatalf("Expected athlete id 42, got %d", athlete.ID)
	}
	if athlete.Username != "runner42" {
		t.Fatalf("Expected username runner42, got %q", athlete.Username)
	}
	if athlete.Firstname != "Ada" {
		t.Fatalf("Expected firstname Ada, got %q", athlete.Firstname)
	}
	if athlete.Lastname != "Runner" {
		t.Fatalf("Expected lastname Runner, got %q", athlete.Lastname)
	}
	if athlete.City != "Berlin" {
		t.Fatalf("Expected city Berlin, got %q", athlete.City)
	}
	if athlete.Country != "Germany" {
		t.Fatalf("Expected country Germany, got %q", athlete.Country)
	}
	if athlete.Sex != "F" {
		t.Fatalf("Expected sex F, got %q", athlete.Sex)
	}
	if !athlete.Premium {
		t.Fatal("Expected premium athlete")
	}
}

func TestFetchStravaAthlete_RequiresAccessToken(t *testing.T) {
	_, err := fetchStravaAthlete("")
	if err == nil {
		t.Fatal("Expected error for missing access token, got nil")
	}
}

func TestFetchStravaAthlete_ReturnsErrorForNonOKStatus(t *testing.T) {
	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
	})
	stravaAPIBaseURL = stravaServer.URL

	_, err := fetchStravaAthlete("invalid-access-token")
	if err == nil {
		t.Fatal("Expected error for non-OK Strava response, got nil")
	}
}

func TestFetchStravaAthlete_ReturnsErrorForInvalidJSON(t *testing.T) {
	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{`))
	}))
	defer stravaServer.Close()

	originalStravaAPIBaseURL := stravaAPIBaseURL
	t.Cleanup(func() {
		stravaAPIBaseURL = originalStravaAPIBaseURL
	})
	stravaAPIBaseURL = stravaServer.URL

	_, err := fetchStravaAthlete("valid-access-token")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestAthleteHandler(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	stravaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":42,"username":"runner42"}`))
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

	req := httptest.NewRequest("GET", "/athlete", nil)
	w := httptest.NewRecorder()

	athleteHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200 OK, got %d: %s", resp.StatusCode, string(body))
	}

	var athlete Athlete
	if err := json.NewDecoder(resp.Body).Decode(&athlete); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if athlete.ID != 42 {
		t.Fatalf("Expected athlete id 42, got %d", athlete.ID)
	}
	if athlete.Username != "runner42" {
		t.Fatalf("Expected username runner42, got %q", athlete.Username)
	}
}

func TestAthleteHandler_ReturnsErrorWhenTokenIsMissing(t *testing.T) {
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		providerTokenStore = originalProviderTokenStore
	})

	providerTokenStore = &fakeTokenStore{err: errTestTokenNotFound}

	req := httptest.NewRequest("GET", "/athlete", nil)
	w := httptest.NewRecorder()

	athleteHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}

func TestAthleteHandler_ReturnsBadGatewayWhenStravaFetchFails(t *testing.T) {
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
			AccessToken:  "valid-access-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
		},
	}

	req := httptest.NewRequest("GET", "/athlete", nil)
	w := httptest.NewRecorder()

	athleteHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("Expected status 502 Bad Gateway, got %d", resp.StatusCode)
	}
}
