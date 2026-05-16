package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fakeTokenStore struct {
	token     ProviderToken
	err       error
	saveCount int
}

func (store *fakeTokenStore) SaveProviderToken(ctx context.Context, token ProviderToken) error {
	store.token = token
	store.saveCount++
	return store.err
}

func (store *fakeTokenStore) GetProviderToken(ctx context.Context, userID string, provider string) (ProviderToken, error) {
	return store.token, store.err
}

type fakeRawObjectStore struct {
	object RawObject
	err    error
}

func (store *fakeRawObjectStore) SaveRawObject(ctx context.Context, object RawObject) error {
	store.object = object
	return store.err
}

type fakeActivityStore struct {
	activities []Activity
	saveCount  int
	err        error
}

func (store *fakeActivityStore) SaveActivities(ctx context.Context, activities []Activity) error {
	store.activities = activities
	store.saveCount++
	return store.err
}

func (store *fakeActivityStore) ListActivities(ctx context.Context, userID string) ([]Activity, error) {
	return store.activities, store.err
}

func TestAuthStravaHandler(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "test-client-id")

	req := httptest.NewRequest("GET", "/auth/strava", nil)
	w := httptest.NewRecorder()

	authStravaHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("Expected status 302 Found, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		t.Fatal("Expected Location header for redirect, got empty")
	}

	redirectUrl, err := url.Parse(location)
	if err != nil {
		t.Fatalf("Invalid redirect URL, got %q: %v", location, err)
	}

	state := redirectUrl.Query().Get("state")
	if state == "" {
		t.Fatal("Expected state parameter in redirect URL, got empty")
	}

	expectedPrefix := "https://www.strava.com/oauth/authorize"
	if len(location) < len(expectedPrefix) || location[:len(expectedPrefix)] != expectedPrefix {
		t.Fatalf("Expected redirect to Strava OAuth, got %s", location)
	}

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("Expected OAuth state cookie, got none")
	}
}

func TestAuthStravaHandler_MissingClientID(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "")

	req := httptest.NewRequest("GET", "/auth/strava", nil)
	w := httptest.NewRecorder()

	authStravaHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error for missing client ID, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "STRAVA_CLIENT_ID is not configured") {
		t.Fatalf("Expected missing client ID error, got %q", string(body))
	}
}

func TestAuthStravaCallbackHandler_MissingCode(t *testing.T) {
	req := httptest.NewRequest("GET", "/auth/strava/callback", nil)
	w := httptest.NewRecorder()

	authStravaCallbackHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request for missing code, got %d", resp.StatusCode)
	}
}

func TestAuthStravaCallbackHandler_MissingCredentials(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "")
	t.Setenv("STRAVA_CLIENT_SECRET", "")

	req := httptest.NewRequest("GET", "/auth/strava/callback?code=somecode&state=test-state", nil)
	req.AddCookie(&http.Cookie{
		Name:  stravaOAuthStateCookieName,
		Value: "test-state",
	})
	w := httptest.NewRecorder()

	authStravaCallbackHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 Internal Server Error for missing credentials, got %d", resp.StatusCode)
	}
}

func TestAuthStravaCallbackHandler_MissingState(t *testing.T) {
	tests := []struct {
		name      string
		targetUrl string
		addCookie bool
	}{
		{
			name:      "Missing URL state and cookie",
			targetUrl: "/auth/strava/callback?code=somecode",
			addCookie: false,
		},
		{
			name:      "Missing URL state but has cookie",
			targetUrl: "/auth/strava/callback?code=somecode",
			addCookie: true,
		},
		{
			name:      "Missing state cookie",
			targetUrl: "/auth/strava/callback?code=somecode&state=test-state",
			addCookie: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.targetUrl, nil)
			if tt.addCookie {
				req.AddCookie(&http.Cookie{
					Name:  stravaOAuthStateCookieName,
					Value: "test-state",
				})
			}

			w := httptest.NewRecorder()

			authStravaCallbackHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("Expected status 400 Bad Request for missing state, got %d", resp.StatusCode)
			}
		})
	}
}

func TestAuthStravaCallbackHandler_InvalidState(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "test-client-id")
	t.Setenv("STRAVA_CLIENT_SECRET", "")

	req := httptest.NewRequest("GET", "/auth/strava/callback?code=somecode&state=invalid-state", nil)
	req.AddCookie(&http.Cookie{
		Name:  stravaOAuthStateCookieName,
		Value: "test-state",
	})
	w := httptest.NewRecorder()

	authStravaCallbackHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request for invalid state, got %d", resp.StatusCode)
	}
}

func TestAuthStravaCallbackHandler_Success(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "test-client-id")
	t.Setenv("STRAVA_CLIENT_SECRET", "test-client-secret")
	t.Setenv("APP_USER_ID", "00000000-0000-0000-0000-000000000042")
	t.Setenv("STRAVA_SCOPES", "activity:read_all")

	originalPostStravaTokenForm := postStravaTokenForm
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		postStravaTokenForm = originalPostStravaTokenForm
		providerTokenStore = originalProviderTokenStore
	})

	tokenStore := &fakeTokenStore{}
	providerTokenStore = tokenStore

	postStravaTokenForm = func(tokenURL string, data url.Values) (*http.Response, error) {
		if tokenURL != "https://www.strava.com/oauth/token" {
			t.Fatalf("Expected Strava token URL, got %s", tokenURL)
		}

		if data.Get("client_id") != "test-client-id" {
			t.Fatalf("Expected client_id test-client-id, got %q", data.Get("client_id"))
		}

		if data.Get("client_secret") != "test-client-secret" {
			t.Fatalf("Expected client_secret test-client-secret, got %q", data.Get("client_secret"))
		}

		if data.Get("code") != "somecode" {
			t.Fatalf("Expected code somecode, got %q", data.Get("code"))
		}

		if data.Get("grant_type") != "authorization_code" {
			t.Fatalf("Expected grant_type authorization_code, got %q", data.Get("grant_type"))
		}

		body := `{
			"access_token": "test-access-token",
			"refresh_token": "test-refresh-token",
			"expires_at": 1234567890,
			"athlete": {
				"id": 42
			}
		}`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}

	req := httptest.NewRequest("GET", "/auth/strava/callback?code=somecode&state=test-state", nil)
	req.AddCookie(&http.Cookie{
		Name:  stravaOAuthStateCookieName,
		Value: "test-state",
	})
	w := httptest.NewRecorder()

	authStravaCallbackHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	var response struct {
		Status    string `json:"status"`
		ExpiresAt int64  `json:"expires_at"`
		AthleteID int64  `json:"athlete_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Status != "connected" {
		t.Fatalf("Expected status connected, got %q", response.Status)
	}

	if response.ExpiresAt != 1234567890 {
		t.Fatalf("Expected expires_at 1234567890, got %d", response.ExpiresAt)
	}

	if response.AthleteID != 42 {
		t.Fatalf("Expected athlete_id 42, got %d", response.AthleteID)
	}

	if tokenStore.token.UserID != "00000000-0000-0000-0000-000000000042" {
		t.Fatalf("Expected stored user ID, got %q", tokenStore.token.UserID)
	}

	if tokenStore.token.Provider != "strava" {
		t.Fatalf("Expected stored provider strava, got %q", tokenStore.token.Provider)
	}

	if tokenStore.token.ProviderUserID != "42" {
		t.Fatalf("Expected stored provider user ID 42, got %q", tokenStore.token.ProviderUserID)
	}

	if tokenStore.token.AccessToken != "test-access-token" {
		t.Fatalf("Expected stored access token, got %q", tokenStore.token.AccessToken)
	}

	if tokenStore.token.RefreshToken != "test-refresh-token" {
		t.Fatalf("Expected stored refresh token, got %q", tokenStore.token.RefreshToken)
	}

	if tokenStore.token.Scopes != "activity:read_all" {
		t.Fatalf("Expected stored scope activity:read_all, got %q", tokenStore.token.Scopes)
	}
}

func TestGetValidStravaToken_ReturnsStoredTokenWhenStillValid(t *testing.T) {
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		providerTokenStore = originalProviderTokenStore
	})

	storedToken := ProviderToken{
		UserID:         "00000000-0000-0000-0000-000000000042",
		Provider:       "strava",
		ProviderUserID: "42",
		AccessToken:    "valid-access-token",
		RefreshToken:   "valid-refresh-token",
		ExpiresAt:      time.Now().UTC().Add(30 * time.Minute),
		Scopes:         "activity:read_all",
	}
	tokenStore := &fakeTokenStore{token: storedToken}
	providerTokenStore = tokenStore

	token, err := getValidStravaToken(context.Background(), storedToken.UserID)
	if err != nil {
		t.Fatalf("Expected stored token, got error: %v", err)
	}

	if token.AccessToken != storedToken.AccessToken {
		t.Fatalf("Expected access token %q, got %q", storedToken.AccessToken, token.AccessToken)
	}

	if tokenStore.saveCount != 0 {
		t.Fatalf("Expected token not to be saved, got %d saves", tokenStore.saveCount)
	}
}

func TestGetValidStravaToken_RefreshesExpiredToken(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "test-client-id")
	t.Setenv("STRAVA_CLIENT_SECRET", "test-client-secret")

	originalPostStravaTokenForm := postStravaTokenForm
	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		postStravaTokenForm = originalPostStravaTokenForm
		providerTokenStore = originalProviderTokenStore
	})

	storedToken := ProviderToken{
		UserID:         "00000000-0000-0000-0000-000000000042",
		Provider:       "strava",
		ProviderUserID: "42",
		AccessToken:    "expired-access-token",
		RefreshToken:   "old-refresh-token",
		ExpiresAt:      time.Now().UTC().Add(-time.Minute),
		Scopes:         "activity:read_all",
	}
	tokenStore := &fakeTokenStore{token: storedToken}
	providerTokenStore = tokenStore

	postStravaTokenForm = func(tokenURL string, data url.Values) (*http.Response, error) {
		if tokenURL != "https://www.strava.com/oauth/token" {
			t.Fatalf("Expected Strava token URL, got %s", tokenURL)
		}

		if data.Get("grant_type") != "refresh_token" {
			t.Fatalf("Expected grant_type refresh_token, got %q", data.Get("grant_type"))
		}

		if data.Get("refresh_token") != "old-refresh-token" {
			t.Fatalf("Expected old refresh token, got %q", data.Get("refresh_token"))
		}

		body := `{
			"access_token": "new-access-token",
			"refresh_token": "new-refresh-token",
			"expires_at": 1893456000
		}`

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}

	token, err := getValidStravaToken(context.Background(), storedToken.UserID)
	if err != nil {
		t.Fatalf("Expected refreshed token, got error: %v", err)
	}

	if token.AccessToken != "new-access-token" {
		t.Fatalf("Expected refreshed access token, got %q", token.AccessToken)
	}

	if token.RefreshToken != "new-refresh-token" {
		t.Fatalf("Expected rotated refresh token, got %q", token.RefreshToken)
	}

	if tokenStore.saveCount != 1 {
		t.Fatalf("Expected refreshed token to be saved once, got %d saves", tokenStore.saveCount)
	}

	if tokenStore.token.RefreshToken != "new-refresh-token" {
		t.Fatalf("Expected saved rotated refresh token, got %q", tokenStore.token.RefreshToken)
	}
}

func TestGetValidStravaToken_RefreshRequiresCredentials(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "")
	t.Setenv("STRAVA_CLIENT_SECRET", "")

	originalProviderTokenStore := providerTokenStore
	t.Cleanup(func() {
		providerTokenStore = originalProviderTokenStore
	})

	tokenStore := &fakeTokenStore{
		token: ProviderToken{
			UserID:       "00000000-0000-0000-0000-000000000042",
			Provider:     "strava",
			AccessToken:  "expired-access-token",
			RefreshToken: "old-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(-time.Minute),
		},
	}
	providerTokenStore = tokenStore

	_, err := getValidStravaToken(context.Background(), tokenStore.token.UserID)
	if err == nil {
		t.Fatal("Expected error for missing credentials, got nil")
	}

	if !strings.Contains(err.Error(), "Strava OAuth credentials are not configured") {
		t.Fatalf("Expected missing credentials error, got %v", err)
	}
}

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

	activitiesHandler(w, req)

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
