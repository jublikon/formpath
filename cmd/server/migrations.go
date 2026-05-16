package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func runMigrations(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading migrations: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, file := range files {
		applied, err := migrationApplied(db, file)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		path := filepath.Join(dir, file)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", file, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("starting migration %s: %w", file, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying migration %s: %w", file, err)
		}

		if _, err := tx.Exec("insert into schema_migrations (version) values ($1)", file); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", file, err)
		}
	}

	return nil
}

func migrationApplied(db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		select exists (
			select 1
			from information_schema.tables
			where table_schema = 'public'
			and table_name = 'schema_migrations'
		)
	`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking schema_migrations table: %w", err)
	}
	if !exists {
		return false, nil
	}

	err = db.QueryRow("select exists (select 1 from schema_migrations where version = $1)", version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking migration %s: %w", version, err)
	}
	return exists, nil
}
