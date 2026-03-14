package store

import (
	"database/sql"
	"fmt"
	"time"
)

type StandupEntry struct {
	ID        int64
	Date      string
	Yesterday string
	Today     string
	Blockers  string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// UpsertStandup creates or updates a standup entry for the given date.
// Uses ON CONFLICT DO UPDATE to preserve the original created_at timestamp.
func UpsertStandup(db *sql.DB, date, yesterday, today, blockers string) error {
	_, err := db.Exec(`
		INSERT INTO standup_entries (date, yesterday, today, blockers, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(date) DO UPDATE SET
			yesterday = excluded.yesterday,
			today = excluded.today,
			blockers = excluded.blockers,
			updated_at = CURRENT_TIMESTAMP
	`, date, yesterday, today, blockers)
	if err != nil {
		return fmt.Errorf("upsert standup: %w", err)
	}
	return nil
}

func GetStandup(db *sql.DB, date string) (*StandupEntry, error) {
	var e StandupEntry
	err := db.QueryRow(
		"SELECT id, CAST(date AS TEXT), COALESCE(yesterday, ''), COALESCE(today, ''), COALESCE(blockers, ''), created_at, updated_at FROM standup_entries WHERE date = ?",
		date,
	).Scan(&e.ID, &e.Date, &e.Yesterday, &e.Today, &e.Blockers, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get standup: %w", err)
	}
	return &e, nil
}

func GetTodayStandup(db *sql.DB) (*StandupEntry, error) {
	today := time.Now().Format("2006-01-02")
	return GetStandup(db, today)
}

func ListStandupHistory(db *sql.DB, limit int) ([]StandupEntry, error) {
	rows, err := db.Query(
		"SELECT id, CAST(date AS TEXT), COALESCE(yesterday, ''), COALESCE(today, ''), COALESCE(blockers, ''), created_at, updated_at FROM standup_entries ORDER BY date DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list standup history: %w", err)
	}
	defer rows.Close()

	var entries []StandupEntry
	for rows.Next() {
		var e StandupEntry
		if err := rows.Scan(&e.ID, &e.Date, &e.Yesterday, &e.Today, &e.Blockers, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan standup: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
