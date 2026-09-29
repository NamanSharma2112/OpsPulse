package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// migrationLockKey serialises migration across instances. Any fixed number
// works; it only has to be the same in every process.
const migrationLockKey int64 = 8_241_557

// Migrate applies every pending .up.sql file in dir, in filename order.
//
// It runs on startup so a deployment is one step: point the service at an
// empty database and it brings itself up. Applied versions are recorded in
// schema_migrations, and an advisory lock keeps two instances booting at once
// from racing each other.
func Migrate(ctx context.Context, db *DB, dir string, log *slog.Logger) error {
	files, err := migrationFiles(dir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .up.sql files found in %q", dir)
	}

	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	// Held until the connection is released, so a crash mid-migration does
	// not leave the lock stuck.
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLockKey); err != nil {
		return fmt.Errorf("take migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, migrationLockKey)
	}()

	var fresh bool
	if err := conn.QueryRow(ctx,
		`SELECT to_regclass('public.schema_migrations') IS NULL`).Scan(&fresh); err != nil {
		return fmt.Errorf("check migration table: %w", err)
	}

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	// A database built by the old compose init script has the full schema but
	// no ledger. That script applied every up migration, so such a database is
	// at head by construction: record the versions rather than replaying them
	// onto tables that already exist.
	if fresh {
		var adopted bool
		if err := conn.QueryRow(ctx,
			`SELECT to_regclass('public.users') IS NOT NULL`).Scan(&adopted); err != nil {
			return fmt.Errorf("check existing schema: %w", err)
		}
		if adopted {
			log.Warn("existing schema found without a migration ledger; " +
				"recording current migrations as applied")
			for _, f := range files {
				if _, err := conn.Exec(ctx,
					`INSERT INTO schema_migrations (version) VALUES ($1)
					 ON CONFLICT (version) DO NOTHING`, version(f)); err != nil {
					return fmt.Errorf("adopt %s: %w", version(f), err)
				}
			}
			return nil
		}
	}

	applied := map[string]bool{}
	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return fmt.Errorf("scan applied migration: %w", err)
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}

	pending := 0
	for _, path := range files {
		v := version(path)
		if applied[v] {
			continue
		}

		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", v, err)
		}

		// Each file already wraps itself in BEGIN/COMMIT, so it is applied as
		// written; the ledger row is inserted separately and only on success.
		if _, err := conn.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("apply %s: %w", v, err)
		}
		if _, err := conn.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, v); err != nil {
			return fmt.Errorf("record %s: %w", v, err)
		}

		log.Info("applied migration", "version", v)
		pending++
	}

	if pending == 0 {
		log.Info("database schema is up to date", "migrations", len(files))
	}
	return nil
}

// migrationFiles returns the .up.sql files in dir, sorted by name.
func migrationFiles(dir string) ([]string, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)
	return entries, nil
}

// version is the filename without its .up.sql suffix, e.g. "0001_init".
func version(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".up.sql")
}

// FindMigrationsDir resolves the migrations directory, trying the configured
// path first and then the layouts used by the container image and by `go run`
// from the service directory.
func FindMigrationsDir(configured string) (string, error) {
	candidates := []string{configured, "/app/migrations", "migrations", "../../infra/migrations"}
	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		if files, err := migrationFiles(dir); err == nil && len(files) > 0 {
			return dir, nil
		}
	}
	return "", fmt.Errorf("no migrations found (looked in %s)",
		strings.Join(filterEmpty(candidates), ", "))
}

func filterEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
