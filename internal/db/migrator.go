package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

type Migrator struct {
	db     *pgxpool.Pool
	logger zerolog.Logger
}

func NewMigrator(db *pgxpool.Pool, logger zerolog.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

func (m *Migrator) Migrate(ctx context.Context) error {
	var tableExists bool
	err := m.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'migrations')`).
		Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !tableExists {
		if _, err := m.db.Exec(ctx,
			`CREATE TABLE migrations (version INT PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT NOW())`); err != nil {
			return fmt.Errorf("failed to create migrations table: %w", err)
		}
	}

	rows, err := m.db.Query(ctx, `SELECT version FROM migrations ORDER BY version DESC`)
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	files, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrations := make(map[int]string)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		versionStr := strings.Split(file.Name(), "_")[0]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("invalid migration file name: %s", file.Name())
		}

		if applied[version] {
			continue
		}

		content, err := fs.ReadFile(migrationsFS, path.Join("migrations", file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		migrations[version] = string(content)
	}

	versions := make([]int, 0, len(migrations))
	for v := range migrations {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	for _, version := range versions {
		m.logger.Info().Int("version", version).Msg("Applying migration")

		tx, err := m.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(ctx, migrations[version]); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %d: %w", version, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO migrations (version) VALUES ($1)`, version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %d: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", version, err)
		}

		m.logger.Info().Int("version", version).Msg("Migration applied successfully")
	}

	return nil
}
