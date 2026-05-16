package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const stravaOAuthStateCookieName = "strava_oauth_state"
const stravaAPITimeout = 10 * time.Second
const stravaTokenRefreshWindow = 5 * time.Minute

var stravaHTTPClient = &http.Client{Timeout: stravaAPITimeout}
var postStravaTokenForm = stravaHTTPClient.PostForm
var providerTokenStore TokenStore = noopTokenStore{}
var providerRawObjectStore RawObjectStore = noopRawObjectStore{}
var providerActivityStore ActivityStore = noopActivityStore{}
var stravaAPIBaseURL = "https://www.strava.com/api/v3"

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

type StravaTokenResponse struct {
	TokenType    string  `json:"token_type"`
	ExpiresAt    int64   `json:"expires_at"`
	ExpiresIn    int64   `json:"expires_in"`
	RefreshToken string  `json:"refresh_token"`
	AccessToken  string  `json:"access_token"`
	Athlete      Athlete `json:"athlete"`
}

func generateOAuthState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func setOAuthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     stravaOAuthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})
}

func clearOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     stravaOAuthStateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func authStravaHandler(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("STRAVA_CLIENT_ID")
	if clientID == "" {
		http.Error(w, "STRAVA_CLIENT_ID is not configured", http.StatusInternalServerError)
		return
	}

	state, err := generateOAuthState()
	if err != nil {
		http.Error(w, "failed to generate OAuth state", http.StatusInternalServerError)
		return
	}

	setOAuthStateCookie(w, state)

	cfg := loadAppConfig()

	u, _ := url.Parse("https://www.strava.com/oauth/authorize")
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", cfg.StravaRedirectURL)
	q.Set("scope", cfg.StravaScopes)
	// OAuth state via short-lived HttpOnly cookie for local-first CSRF protection.
	q.Set("state", state)
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}

func authStravaCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	returnedState := r.URL.Query().Get("state")
	if returnedState == "" {
		http.Error(w, "missing OAuth state", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie(stravaOAuthStateCookieName)
	if err != nil {
		http.Error(w, "missing OAuth state cookie", http.StatusBadRequest)
		return
	}

	if returnedState != cookie.Value {
		http.Error(w, "invalid OAuth state", http.StatusBadRequest)
		return
	}

	clearOAuthStateCookie(w)

	clientID := os.Getenv("STRAVA_CLIENT_ID")
	clientSecret := os.Getenv("STRAVA_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		http.Error(w, "Strava OAuth credentials are not configured", http.StatusInternalServerError)
		return
	}

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")

	resp, err := postStravaTokenForm("https://www.strava.com/oauth/token", form)
	if err != nil {
		http.Error(w, "failed to exchange authorization code", http.StatusBadGateway)
		return
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		http.Error(w, "Strava token exchange failed", http.StatusBadGateway)
		return
	}

	body, err := readAllAndClose(resp.Body)
	if err != nil {
		http.Error(w, "failed to read Strava token response", http.StatusBadGateway)
		return
	}

	var tokenResponse StravaTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		http.Error(w, "failed to decode Strava token response", http.StatusBadGateway)
		return
	}

	if tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" || tokenResponse.ExpiresAt == 0 {
		http.Error(w, "Strava token response is incomplete", http.StatusBadGateway)
		return
	}

	cfg := loadAppConfig()
	err = providerTokenStore.SaveProviderToken(r.Context(), ProviderToken{
		UserID:         cfg.AppUserID,
		Provider:       "strava",
		ProviderUserID: strconv.FormatInt(tokenResponse.Athlete.ID, 10),
		AccessToken:    tokenResponse.AccessToken,
		RefreshToken:   tokenResponse.RefreshToken,
		ExpiresAt:      time.Unix(tokenResponse.ExpiresAt, 0).UTC(),
		Scopes:         cfg.StravaScopes,
	})
	if err != nil {
		http.Error(w, "failed to store Strava tokens", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status    string `json:"status"`
		ExpiresAt int64  `json:"expires_at"`
		AthleteID int64  `json:"athlete_id"`
	}{
		Status:    "connected",
		ExpiresAt: tokenResponse.ExpiresAt,
		AthleteID: tokenResponse.Athlete.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func refreshStravaToken(token ProviderToken) (ProviderToken, error) {
	if token.RefreshToken == "" {
		return ProviderToken{}, errors.New("refresh token is required")
	}

	clientID := os.Getenv("STRAVA_CLIENT_ID")
	clientSecret := os.Getenv("STRAVA_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return ProviderToken{}, errors.New("Strava OAuth credentials are not configured")
	}

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("refresh_token", token.RefreshToken)
	form.Set("grant_type", "refresh_token")

	resp, err := postStravaTokenForm("https://www.strava.com/oauth/token", form)
	if err != nil {
		return ProviderToken{}, fmt.Errorf("refreshing Strava token: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return ProviderToken{}, fmt.Errorf("Strava token refresh failed with status %d", resp.StatusCode)
	}

	body, err := readAllAndClose(resp.Body)
	if err != nil {
		return ProviderToken{}, fmt.Errorf("reading Strava refresh response: %w", err)
	}

	var tokenResponse StravaTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return ProviderToken{}, fmt.Errorf("decoding Strava refresh response: %w", err)
	}

	if tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" || tokenResponse.ExpiresAt == 0 {
		return ProviderToken{}, errors.New("Strava refresh response is incomplete")
	}

	refreshed := token
	refreshed.AccessToken = tokenResponse.AccessToken
	refreshed.RefreshToken = tokenResponse.RefreshToken
	refreshed.ExpiresAt = time.Unix(tokenResponse.ExpiresAt, 0).UTC()
	return refreshed, nil
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

func getValidStravaToken(ctx context.Context, userID string) (ProviderToken, error) {
	token, err := providerTokenStore.GetProviderToken(ctx, userID, "strava")
	if err != nil {
		return ProviderToken{}, fmt.Errorf("loading Strava token: %w", err)
	}

	if time.Now().UTC().Before(token.ExpiresAt.Add(-stravaTokenRefreshWindow)) {
		return token, nil
	}

	refreshed, err := refreshStravaToken(token)
	if err != nil {
		return ProviderToken{}, err
	}

	if err := providerTokenStore.SaveProviderToken(ctx, refreshed); err != nil {
		return ProviderToken{}, fmt.Errorf("saving refreshed Strava token: %w", err)
	}

	return refreshed, nil
}
