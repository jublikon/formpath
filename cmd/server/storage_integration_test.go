package main

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func openIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	if os.Getenv("FORMPATH_DB_TEST") != "1" {
		t.Skip("set FORMPATH_DB_TEST=1 to run Postgres integration tests")
	}

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL or DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connecting to database: %v", err)
	}
	if err := runMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	return db
}

func cleanupIntegrationUser(t *testing.T, db *sql.DB, userID string) {
	t.Helper()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := db.ExecContext(ctx, "delete from activities where user_id = $1", userID); err != nil {
			t.Errorf("cleaning activities: %v", err)
		}
		if _, err := db.ExecContext(ctx, "delete from raw_objects where user_id = $1", userID); err != nil {
			t.Errorf("cleaning raw objects: %v", err)
		}
		if _, err := db.ExecContext(ctx, "delete from provider_tokens where user_id = $1", userID); err != nil {
			t.Errorf("cleaning provider tokens: %v", err)
		}
	})
}

func TestPostgresTokenStore_SaveAndUpdateProviderToken(t *testing.T) {
	db := openIntegrationDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID := "00000000-0000-0000-0000-000000000101"
	cleanupIntegrationUser(t, db, userID)

	store := NewPostgresTokenStore(db)
	token := ProviderToken{
		UserID:         userID,
		Provider:       "strava",
		ProviderUserID: "42",
		AccessToken:    "first-access-token",
		RefreshToken:   "first-refresh-token",
		ExpiresAt:      time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC),
		Scopes:         "activity:read_all",
	}
	if err := store.SaveProviderToken(ctx, token); err != nil {
		t.Fatalf("saving token: %v", err)
	}

	token.AccessToken = "second-access-token"
	token.RefreshToken = "second-refresh-token"
	token.ExpiresAt = time.Date(2026, 5, 17, 11, 0, 0, 0, time.UTC)
	if err := store.SaveProviderToken(ctx, token); err != nil {
		t.Fatalf("updating token: %v", err)
	}

	got, err := store.GetProviderToken(ctx, userID, "strava")
	if err != nil {
		t.Fatalf("loading token: %v", err)
	}

	if got.AccessToken != "second-access-token" {
		t.Fatalf("Expected updated access token, got %q", got.AccessToken)
	}
	if got.RefreshToken != "second-refresh-token" {
		t.Fatalf("Expected updated refresh token, got %q", got.RefreshToken)
	}
	if got.Scopes != "activity:read_all" {
		t.Fatalf("Expected stored scopes, got %q", got.Scopes)
	}
}

func TestPostgresActivityStore_DeduplicatesAndListsActivities(t *testing.T) {
	db := openIntegrationDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID := "00000000-0000-0000-0000-000000000102"
	cleanupIntegrationUser(t, db, userID)

	store := NewPostgresActivityStore(db)
	elevationGain := 42.5
	activity := Activity{
		UserID:              userID,
		Provider:            "strava",
		ProviderID:          "integration-activity-102",
		ActivityType:        "run",
		Name:                "First Name",
		StartedAt:           time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC),
		DurationSeconds:     3600,
		MovingSeconds:       3500,
		DistanceMeters:      10000,
		ElevationGainMeters: &elevationGain,
		RawObjectKey:        "raw/object/first.json",
	}
	if err := store.SaveActivities(ctx, []Activity{activity}); err != nil {
		t.Fatalf("saving activity: %v", err)
	}

	activity.Name = "Updated Name"
	activity.DistanceMeters = 10050
	activity.ElevationGainMeters = nil
	activity.RawObjectKey = ""
	if err := store.SaveActivities(ctx, []Activity{activity}); err != nil {
		t.Fatalf("updating activity: %v", err)
	}

	activities, err := store.ListActivities(ctx, userID)
	if err != nil {
		t.Fatalf("listing activities: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("Expected one deduplicated activity, got %d", len(activities))
	}

	got := activities[0]
	if got.Name != "Updated Name" {
		t.Fatalf("Expected updated name, got %q", got.Name)
	}
	if got.DistanceMeters != 10050 {
		t.Fatalf("Expected updated distance, got %f", got.DistanceMeters)
	}
	if got.ElevationGainMeters != nil {
		t.Fatalf("Expected nil elevation gain after update, got %v", got.ElevationGainMeters)
	}
	if got.RawObjectKey != "" {
		t.Fatalf("Expected empty raw object key after update, got %q", got.RawObjectKey)
	}
	if got.ID == "" {
		t.Fatal("Expected generated activity id")
	}
	if got.CreatedAt == nil || got.UpdatedAt == nil {
		t.Fatal("Expected created_at and updated_at to be loaded")
	}
}

func TestRunMigrations_IsIdempotent(t *testing.T) {
	db := openIntegrationDB(t)

	if err := runMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("running migrations second time: %v", err)
	}
}

func TestMinIORawObjectStore_SaveRawObjectRecordsMetadata(t *testing.T) {
	if os.Getenv("FORMPATH_S3_TEST") != "1" {
		t.Skip("set FORMPATH_S3_TEST=1 to run MinIO integration tests")
	}

	db := openIntegrationDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID := "00000000-0000-0000-0000-000000000103"
	cleanupIntegrationUser(t, db, userID)

	cfg := appConfig{
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3Bucket:          envOrDefault("S3_BUCKET", "formpath-raw-test"),
		S3UseSSL:          strings.EqualFold(os.Getenv("S3_USE_SSL"), "true"),
	}
	if cfg.S3Endpoint == "" || cfg.S3AccessKeyID == "" || cfg.S3SecretAccessKey == "" {
		t.Fatal("S3_ENDPOINT, S3_ACCESS_KEY_ID, and S3_SECRET_ACCESS_KEY are required")
	}

	store, err := NewMinIORawObjectStore(ctx, db, cfg)
	if err != nil {
		t.Fatalf("creating raw object store: %v", err)
	}

	object := RawObject{
		UserID:             userID,
		Provider:           "strava",
		ProviderObjectType: "activity_list",
		ProviderObjectID:   "integration-activity-list-103",
		ObjectKey:          "integration/activity-list-103.json",
		ContentType:        "application/json",
		Body:               []byte(`[{"id":123}]`),
		FetchedAt:          time.Date(2026, 5, 17, 9, 0, 0, 0, time.UTC),
	}
	if err := store.SaveRawObject(ctx, object); err != nil {
		t.Fatalf("saving raw object: %v", err)
	}

	var sha string
	var sizeBytes int
	var contentType string
	err = db.QueryRowContext(ctx, `
		select sha256, size_bytes, content_type
		from raw_objects
		where user_id = $1
		and provider = $2
		and provider_object_type = $3
		and provider_object_id = $4
	`, object.UserID, object.Provider, object.ProviderObjectType, object.ProviderObjectID).Scan(&sha, &sizeBytes, &contentType)
	if err != nil {
		t.Fatalf("loading raw object metadata: %v", err)
	}
	if sha != "5cf8edee7646cb9bd5e63244f9b6e1c6e080a7f5aa8196fba47f002c60bcdb19" {
		t.Fatalf("Expected raw object sha256, got %q", sha)
	}
	if sizeBytes != len(object.Body) {
		t.Fatalf("Expected size %d, got %d", len(object.Body), sizeBytes)
	}
	if contentType != "application/json" {
		t.Fatalf("Expected content type application/json, got %q", contentType)
	}
}
