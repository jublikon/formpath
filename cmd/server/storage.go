package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ProviderToken struct {
	UserID         string
	Provider       string
	ProviderUserID string
	AccessToken    string
	RefreshToken   string
	ExpiresAt      time.Time
	Scopes         string
}

type TokenStore interface {
	SaveProviderToken(ctx context.Context, token ProviderToken) error
	GetProviderToken(ctx context.Context, userID string, provider string) (ProviderToken, error)
}

type noopTokenStore struct{}

func (noopTokenStore) SaveProviderToken(ctx context.Context, token ProviderToken) error {
	return nil
}

func (noopTokenStore) GetProviderToken(ctx context.Context, userID string, provider string) (ProviderToken, error) {
	return ProviderToken{}, fmt.Errorf("provider token store is not configured")
}

type PostgresTokenStore struct {
	db *sql.DB
}

func NewPostgresTokenStore(db *sql.DB) *PostgresTokenStore {
	return &PostgresTokenStore{db: db}
}

func (store *PostgresTokenStore) SaveProviderToken(ctx context.Context, token ProviderToken) error {
	_, err := store.db.ExecContext(ctx, `
		insert into provider_tokens (
			user_id,
			provider,
			provider_user_id,
			access_token,
			refresh_token,
			expires_at,
			scopes,
			updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, now())
		on conflict (user_id, provider)
		do update set
			provider_user_id = excluded.provider_user_id,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			expires_at = excluded.expires_at,
			scopes = excluded.scopes,
			updated_at = now()
	`, token.UserID, token.Provider, token.ProviderUserID, token.AccessToken, token.RefreshToken, token.ExpiresAt, token.Scopes)
	if err != nil {
		return fmt.Errorf("saving provider token: %w", err)
	}
	return nil
}

func (store *PostgresTokenStore) GetProviderToken(ctx context.Context, userID string, provider string) (ProviderToken, error) {
	var token ProviderToken
	err := store.db.QueryRowContext(ctx, `
		select
			user_id::text,
			provider,
			provider_user_id,
			access_token,
			refresh_token,
			expires_at,
			scopes
		from provider_tokens
		where user_id = $1
		and provider = $2
	`, userID, provider).Scan(
		&token.UserID,
		&token.Provider,
		&token.ProviderUserID,
		&token.AccessToken,
		&token.RefreshToken,
		&token.ExpiresAt,
		&token.Scopes,
	)
	if err != nil {
		return ProviderToken{}, fmt.Errorf("loading provider token: %w", err)
	}
	return token, nil
}

type RawObject struct {
	UserID             string
	Provider           string
	ProviderObjectType string
	ProviderObjectID   string
	ObjectKey          string
	ContentType        string
	Body               []byte
	FetchedAt          time.Time
}

type RawObjectStore interface {
	SaveRawObject(ctx context.Context, object RawObject) error
}

type noopRawObjectStore struct{}

func (noopRawObjectStore) SaveRawObject(ctx context.Context, object RawObject) error {
	return nil
}

type MinIORawObjectStore struct {
	db     *sql.DB
	client *minio.Client
	bucket string
}

func NewMinIORawObjectStore(ctx context.Context, db *sql.DB, cfg appConfig) (*MinIORawObjectStore, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKeyID, cfg.S3SecretAccessKey, ""),
		Secure: cfg.S3UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("creating MinIO client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.S3Bucket)
	if err != nil {
		return nil, fmt.Errorf("checking raw bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("creating raw bucket: %w", err)
		}
	}

	return &MinIORawObjectStore{
		db:     db,
		client: client,
		bucket: cfg.S3Bucket,
	}, nil
}

func (store *MinIORawObjectStore) SaveRawObject(ctx context.Context, object RawObject) error {
	reader := bytes.NewReader(object.Body)
	_, err := store.client.PutObject(ctx, store.bucket, object.ObjectKey, reader, int64(len(object.Body)), minio.PutObjectOptions{
		ContentType: object.ContentType,
	})
	if err != nil {
		return fmt.Errorf("uploading raw object: %w", err)
	}

	sum := sha256.Sum256(object.Body)
	sha := hex.EncodeToString(sum[:])
	fetchedAt := object.FetchedAt
	if fetchedAt.IsZero() {
		fetchedAt = time.Now().UTC()
	}

	_, err = store.db.ExecContext(ctx, `
		insert into raw_objects (
			user_id,
			provider,
			provider_object_type,
			provider_object_id,
			object_key,
			sha256,
			content_type,
			size_bytes,
			fetched_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		on conflict (provider, provider_object_type, provider_object_id)
		do update set
			object_key = excluded.object_key,
			sha256 = excluded.sha256,
			content_type = excluded.content_type,
			size_bytes = excluded.size_bytes,
			fetched_at = excluded.fetched_at
	`, object.UserID, object.Provider, object.ProviderObjectType, object.ProviderObjectID, object.ObjectKey, sha, object.ContentType, len(object.Body), fetchedAt)
	if err != nil {
		return fmt.Errorf("recording raw object metadata: %w", err)
	}

	return nil
}

type Activity struct {
	ID                  string     `json:"id,omitempty"`
	UserID              string     `json:"user_id"`
	Provider            string     `json:"provider"`
	ProviderID          string     `json:"provider_id"`
	ActivityType        string     `json:"activity_type"`
	Name                string     `json:"name"`
	StartedAt           time.Time  `json:"started_at"`
	DurationSeconds     int        `json:"duration_seconds"`
	MovingSeconds       int        `json:"moving_seconds"`
	DistanceMeters      float64    `json:"distance_meters"`
	ElevationGainMeters *float64   `json:"elevation_gain_meters,omitempty"`
	AverageHeartRateBPM *float64   `json:"average_heartrate_bpm,omitempty"`
	MaxHeartRateBPM     *float64   `json:"max_heartrate_bpm,omitempty"`
	CaloriesKcal        *float64   `json:"calories_kcal,omitempty"`
	RawObjectKey        string     `json:"raw_object_key,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

type ActivityStore interface {
	SaveActivities(ctx context.Context, activities []Activity) error
	ListActivities(ctx context.Context, userID string) ([]Activity, error)
}

type noopActivityStore struct{}

func (noopActivityStore) SaveActivities(ctx context.Context, activities []Activity) error {
	return nil
}

func (noopActivityStore) ListActivities(ctx context.Context, userID string) ([]Activity, error) {
	return nil, fmt.Errorf("activity store is not configured")
}

type PostgresActivityStore struct {
	db *sql.DB
}

func NewPostgresActivityStore(db *sql.DB) *PostgresActivityStore {
	return &PostgresActivityStore{db: db}
}

func (store *PostgresActivityStore) SaveActivities(ctx context.Context, activities []Activity) error {
	if len(activities) == 0 {
		return nil
	}

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting activity transaction: %w", err)
	}
	defer tx.Rollback()

	for _, activity := range activities {
		_, err := tx.ExecContext(ctx, `
			insert into activities (
				user_id,
				provider,
				provider_id,
				activity_type,
				name,
				started_at,
				duration_seconds,
				moving_seconds,
				distance_meters,
				elevation_gain_meters,
				average_heartrate_bpm,
				max_heartrate_bpm,
				calories_kcal,
				raw_object_key,
				updated_at
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now())
			on conflict (provider, provider_id)
			do update set
				user_id = excluded.user_id,
				activity_type = excluded.activity_type,
				name = excluded.name,
				started_at = excluded.started_at,
				duration_seconds = excluded.duration_seconds,
				moving_seconds = excluded.moving_seconds,
				distance_meters = excluded.distance_meters,
				elevation_gain_meters = excluded.elevation_gain_meters,
				average_heartrate_bpm = excluded.average_heartrate_bpm,
				max_heartrate_bpm = excluded.max_heartrate_bpm,
				calories_kcal = excluded.calories_kcal,
				raw_object_key = excluded.raw_object_key,
				updated_at = now()
		`, activity.UserID, activity.Provider, activity.ProviderID, activity.ActivityType, activity.Name, activity.StartedAt, activity.DurationSeconds, activity.MovingSeconds, activity.DistanceMeters, nullableFloat(activity.ElevationGainMeters), nullableFloat(activity.AverageHeartRateBPM), nullableFloat(activity.MaxHeartRateBPM), nullableFloat(activity.CaloriesKcal), nullableString(activity.RawObjectKey))
		if err != nil {
			return fmt.Errorf("saving activity %s/%s: %w", activity.Provider, activity.ProviderID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing activity transaction: %w", err)
	}
	return nil
}

func (store *PostgresActivityStore) ListActivities(ctx context.Context, userID string) ([]Activity, error) {
	rows, err := store.db.QueryContext(ctx, `
		select
			id::text,
			user_id::text,
			provider,
			provider_id,
			activity_type,
			name,
			started_at,
			duration_seconds,
			moving_seconds,
			distance_meters,
			elevation_gain_meters,
			average_heartrate_bpm,
			max_heartrate_bpm,
			calories_kcal,
			coalesce(raw_object_key, ''),
			created_at,
			updated_at
		from activities
		where user_id = $1
		order by started_at desc
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing activities: %w", err)
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var activity Activity
		var elevationGain sql.NullFloat64
		var averageHeartRate sql.NullFloat64
		var maxHeartRate sql.NullFloat64
		var calories sql.NullFloat64
		var createdAt time.Time
		var updatedAt time.Time

		err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Provider,
			&activity.ProviderID,
			&activity.ActivityType,
			&activity.Name,
			&activity.StartedAt,
			&activity.DurationSeconds,
			&activity.MovingSeconds,
			&activity.DistanceMeters,
			&elevationGain,
			&averageHeartRate,
			&maxHeartRate,
			&calories,
			&activity.RawObjectKey,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning activity: %w", err)
		}

		activity.ElevationGainMeters = floatPointerFromNull(elevationGain)
		activity.AverageHeartRateBPM = floatPointerFromNull(averageHeartRate)
		activity.MaxHeartRateBPM = floatPointerFromNull(maxHeartRate)
		activity.CaloriesKcal = floatPointerFromNull(calories)
		activity.CreatedAt = &createdAt
		activity.UpdatedAt = &updatedAt
		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating activities: %w", err)
	}
	return activities, nil
}

func nullableFloat(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func floatPointerFromNull(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}

func readAllAndClose(reader io.ReadCloser) ([]byte, error) {
	defer reader.Close()
	return io.ReadAll(reader)
}
