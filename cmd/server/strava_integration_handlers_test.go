package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStravaIntegrationHandler_ReturnsConnectedWhenRefreshTokenExists(t *testing.T) {
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")

	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		providerTokenStore = originalProviderTokenStore
	})

	providerTokenStore = &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			RefreshToken: "refresh-token",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/integrations/strava", nil)
	w := httptest.NewRecorder()

	stravaIntegrationHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	var response stravaIntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Provider != "strava" {
		t.Fatalf("Expected provider strava, got %q", response.Provider)
	}
	if !response.Connected {
		t.Fatal("Expected Strava integration to be connected")
	}
}

func TestStravaIntegrationHandler_ReturnsDisconnectedWhenTokenIsMissing(t *testing.T) {
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		providerTokenStore = originalProviderTokenStore
	})

	providerTokenStore = &fakeTokenStore{err: sql.ErrNoRows}

	req := httptest.NewRequest(http.MethodGet, "/api/integrations/strava", nil)
	w := httptest.NewRecorder()

	stravaIntegrationHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	var response stravaIntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Provider != "strava" {
		t.Fatalf("Expected provider strava, got %q", response.Provider)
	}
	if response.Connected {
		t.Fatal("Expected Strava integration to be disconnected")
	}
}

func TestStravaIntegrationHandler_ReturnsServerErrorWhenStoreFails(t *testing.T) {
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		providerTokenStore = originalProviderTokenStore
	})

	providerTokenStore = &fakeTokenStore{err: errors.New("database unavailable")}

	req := httptest.NewRequest(http.MethodGet, "/api/integrations/strava", nil)
	w := httptest.NewRecorder()

	stravaIntegrationHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}
