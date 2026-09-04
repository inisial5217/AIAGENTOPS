package repository

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrator runs migrations
type Migrator struct {
	pool    *pgxpool.Pool
	dir     string
	logger  *slog.Logger
}

// NewMigrator creates migrator
func NewMigrator(pool *pgxpool.Pool, dir string, logger *slog.Logger) *Migrator {
	return &Migrator{
		pool:   pool,
		dir:    dir,
		logger: logger,
	}
}

// Up runs migrations
func (m *Migrator) Up(ctx context.Context) error {
	// create schema table
	createSchemaSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	if _, err := m.pool.Exec(ctx, createSchemaSQL); err != nil {
		return fmt.Errorf("create schema table: %w", err)
	}

	// list migration files
	files, err := os.ReadDir(m.dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var upFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	for _, fileName := range upFiles {
		version := strings.TrimSuffix(fileName, ".up.sql")

		// check if applied
		var exists bool
		err := m.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check version %s: %w", version, err)
		}

		if exists {
			continue
		}

		// read migration sql
		filePath := filepath.Join(m.dir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read sql %s: %w", fileName, err)
		}

		// execute migration tx
		tx, err := m.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("exec sql %s: %w", fileName, err)
		}

		insertSQL := "INSERT INTO schema_migrations (version) VALUES ($1)"
		if _, err := tx.Exec(ctx, insertSQL, version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit tx %s: %w", version, err)
		}

		if m.logger != nil {
			m.logger.Info("applied migration", slog.String("version", version))
		}
	}

	return nil
}
