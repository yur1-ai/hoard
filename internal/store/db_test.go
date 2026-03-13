package store

import (
	"testing"
)

func TestOpenInMemory(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer db.Close()

	tables := []string{
		"accounts", "holdings", "transactions",
		"watchlists", "watchlist_items", "market_cache",
		"currency_rates", "tasks", "standup_entries",
		"calendar_cache", "schema_migrations",
	}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("table %s not found: %v", table, err)
		}
	}
}

func TestMigrationIdempotent(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	defer db.Close()

	// Run migrations again — should not error
	if err := RunMigrations(db); err != nil {
		t.Fatalf("second migration run failed: %v", err)
	}
}

func TestMigrationVersionTracking(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		t.Fatalf("query version: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}

	// Run again — should be idempotent (no re-execution)
	if err := RunMigrations(db); err != nil {
		t.Fatalf("second run: %v", err)
	}

	// Version should still be 1, not 2
	err = db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		t.Fatalf("query version after second run: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1 after idempotent run, got %d", version)
	}
}

func TestTaskStatusConstraint(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// 'todo' and 'done' should work
	_, err = db.Exec("INSERT INTO tasks (title, status) VALUES ('test', 'todo')")
	if err != nil {
		t.Errorf("insert 'todo' status failed: %v", err)
	}
	_, err = db.Exec("INSERT INTO tasks (title, status) VALUES ('test2', 'done')")
	if err != nil {
		t.Errorf("insert 'done' status failed: %v", err)
	}

	// 'doing' should be rejected (v3 fix: removed)
	_, err = db.Exec("INSERT INTO tasks (title, status) VALUES ('test3', 'doing')")
	if err == nil {
		t.Error("expected error for 'doing' status, got nil")
	}
}

func TestWALCheckpoint(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := WALCheckpoint(db); err != nil {
		t.Errorf("WALCheckpoint failed: %v", err)
	}
}
