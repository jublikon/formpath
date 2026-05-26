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
