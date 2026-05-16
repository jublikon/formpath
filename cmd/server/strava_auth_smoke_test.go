package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func TestStravaSmoke(t *testing.T) {
	if os.Getenv("STRAVA_SMOKE_TEST") != "1" {
		t.Skip("set STRAVA_SMOKE_TEST=1 to run")
	}

	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("loading .env: %v", err)
	}

	cfg := loadAppConfig()
	if cfg.DatabaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tokenStore := NewPostgresTokenStore(db)
	token, err := tokenStore.GetProviderToken(ctx, cfg.AppUserID, "strava")
	if err != nil {
		t.Fatalf("loading Strava token from database: %v", err)
	}

	athlete, err := fetchStravaAthlete(token.AccessToken)
	if err != nil {
		t.Fatalf("Error fetching Strava athlete: %v", err)
	}

	t.Logf("Athlete: %+v", athlete)
}
