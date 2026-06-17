package main

import (
	"os"
	"strings"
)

const defaultLocalUserID = "00000000-0000-0000-0000-000000000001"

type appConfig struct {
	DatabaseURL       string
	FrontendURL       string
	AppUserID         string
	StravaRedirectURL string
	StravaScopes      string
	S3Endpoint        string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3Bucket          string
	S3UseSSL          bool
}

func loadAppConfig() appConfig {
	return appConfig{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		FrontendURL:       envOrDefault("FRONTEND_URL", "http://localhost:5173"),
		AppUserID:         envOrDefault("APP_USER_ID", defaultLocalUserID),
		StravaRedirectURL: envOrDefault("STRAVA_REDIRECT_URL", "http://localhost:8080/auth/strava/callback"),
		StravaScopes:      envOrDefault("STRAVA_SCOPES", "activity:read_all"),
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3Bucket:          envOrDefault("S3_BUCKET", "formpath-raw"),
		S3UseSSL:          strings.EqualFold(os.Getenv("S3_USE_SSL"), "true"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
