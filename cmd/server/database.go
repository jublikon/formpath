package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const databaseConnectTimeout = 15 * time.Second

type persistenceResources struct {
	db *sql.DB
}

func configurePersistence(cfg appConfig) (*persistenceResources, error) {
	if cfg.DatabaseURL == "" {
		return &persistenceResources{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseConnectTimeout)
	defer cancel()

	db, err := openDatabase(ctx, cfg)
	if err != nil {
		return nil, err
	}

	resources := &persistenceResources{db: db}
	providerTokenStore = NewPostgresTokenStore(db)
	providerActivityStore = NewPostgresActivityStore(db)

	if rawObjectStoreConfigured(cfg) {
		rawStore, err := NewMinIORawObjectStore(ctx, db, cfg)
		if err != nil {
			resources.Close()
			return nil, fmt.Errorf("initializing raw object store: %w", err)
		}
		providerRawObjectStore = rawStore
	}

	return resources, nil
}

func (resources *persistenceResources) Close() error {
	if resources == nil || resources.db == nil {
		return nil
	}
	return resources.db.Close()
}

func openDatabase(ctx context.Context, cfg appConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	if err := runMigrations(db, resolveMigrationsDir()); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

func resolveMigrationsDir() string {
	dir := envOrDefault("MIGRATIONS_DIR", "migrations")
	if _, err := os.Stat(dir); err != nil {
		return "/app/migrations"
	}
	return dir
}

func rawObjectStoreConfigured(cfg appConfig) bool {
	return cfg.S3Endpoint != "" && cfg.S3AccessKeyID != "" && cfg.S3SecretAccessKey != ""
}
