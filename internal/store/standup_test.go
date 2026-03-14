package store

import (
	"testing"
	"time"
)

func TestUpsertStandup(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// Create new entry
	if err := UpsertStandup(db, "2026-03-13", "did A", "will B", "none"); err != nil {
		t.Fatalf("upsert create: %v", err)
	}

	entry, err := GetStandup(db, "2026-03-13")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.Yesterday != "did A" {
		t.Errorf("expected 'did A', got %s", entry.Yesterday)
	}

	// Update same date
	if err := UpsertStandup(db, "2026-03-13", "did C", "will D", "blocked"); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	entry, _ = GetStandup(db, "2026-03-13")
	if entry.Yesterday != "did C" {
		t.Errorf("expected 'did C' after update, got %s", entry.Yesterday)
	}
	if entry.Blockers != "blocked" {
		t.Errorf("expected 'blocked', got %s", entry.Blockers)
	}
}

func TestUpsertPreservesCreatedAt(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	UpsertStandup(db, "2026-03-13", "did A", "will B", "none")
	entry1, _ := GetStandup(db, "2026-03-13")
	createdAt := entry1.CreatedAt

	time.Sleep(10 * time.Millisecond)

	UpsertStandup(db, "2026-03-13", "did C", "will D", "blocked")
	entry2, _ := GetStandup(db, "2026-03-13")

	if !entry2.CreatedAt.Equal(createdAt) {
		t.Errorf("created_at changed: was %v, now %v", createdAt, entry2.CreatedAt)
	}
	if entry2.Yesterday != "did C" {
		t.Errorf("expected 'did C', got %s", entry2.Yesterday)
	}
	if entry2.UpdatedAt == nil || entry2.UpdatedAt.Before(createdAt) {
		t.Error("updated_at should be set and after created_at")
	}
}

func TestGetTodayStandup(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// No entry yet
	entry, err := GetTodayStandup(db)
	if err != nil {
		t.Fatalf("get today empty: %v", err)
	}
	if entry != nil {
		t.Error("expected nil for today with no entry")
	}

	// Add today's entry
	today := time.Now().Format("2006-01-02")
	UpsertStandup(db, today, "yesterday", "today", "none")

	entry, err = GetTodayStandup(db)
	if err != nil {
		t.Fatalf("get today: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry for today")
	}
	if entry.Today != "today" {
		t.Errorf("expected 'today', got %s", entry.Today)
	}
}

func TestListStandupHistory(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	UpsertStandup(db, "2026-03-11", "y1", "t1", "b1")
	UpsertStandup(db, "2026-03-12", "y2", "t2", "b2")
	UpsertStandup(db, "2026-03-13", "y3", "t3", "b3")

	// Limit 2
	entries, err := ListStandupHistory(db, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Ordered by date desc — newest first
	if entries[0].Date != "2026-03-13" {
		t.Errorf("expected newest first (2026-03-13), got %s", entries[0].Date)
	}
	if entries[1].Date != "2026-03-12" {
		t.Errorf("expected second (2026-03-12), got %s", entries[1].Date)
	}
}
