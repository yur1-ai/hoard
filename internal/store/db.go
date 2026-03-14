package store

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/yur1-ai/hoard/migrations"
	_ "modernc.org/sqlite"
)

// Open opens a SQLite database and runs migrations.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// SQLite with database/sql pool: limit to 1 connection to prevent
	// SQLITE_BUSY errors. WAL mode allows concurrent readers on separate
	// OS-level connections but only one writer; with a single pool conn
	// all access is serialized through Go, which is safe for a TUI app.
	db.SetMaxOpenConns(1)

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("pragma %s: %w", p, err)
		}
	}

	if err := RunMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return db, nil
}

// RunMigrations applies embedded SQL migration files with version tracking.
func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	var currentVersion int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	// Parse and sort migrations by numeric version to avoid relying on
	// lexicographic order (which breaks if zero-padding is inconsistent).
	type migration struct {
		version int
		name    string
	}
	var pending []migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		versionStr := strings.SplitN(name, "_", 2)[0]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			slog.Warn("skipping non-versioned migration file", "file", name)
			continue
		}
		if version <= currentVersion {
			continue
		}
		pending = append(pending, migration{version: version, name: name})
	}
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].version < pending[j].version
	})

	for _, m := range pending {
		data, err := migrations.FS.ReadFile(m.name)
		if err != nil {
			return fmt.Errorf("read %s: %w", m.name, err)
		}

		// Wrap each migration + version record in a transaction so
		// partial failures don't leave the DB in an inconsistent state.
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec(string(data)); err != nil {
			tx.Rollback()
			return fmt.Errorf("exec %s: %w", m.name, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
		slog.Info("migration applied", "file", m.name, "version", m.version)
	}
	return nil
}

// WALCheckpoint runs a WAL checkpoint to prevent unbounded WAL growth.
func WALCheckpoint(db *sql.DB) error {
	_, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}
