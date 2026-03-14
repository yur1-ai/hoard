package store

import (
	"testing"
)

func TestCreateWatchlist(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	id, err := CreateWatchlist(db, "Tech Stocks")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestAddWatchlistItem(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	wID, _ := CreateWatchlist(db, "Tech")
	if err := AddWatchlistItem(db, wID, "AAPL", "us_equity"); err != nil {
		t.Fatalf("add item: %v", err)
	}

	items, err := ListWatchlistItems(db, wID)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", items[0].Symbol)
	}
}

func TestListWatchlists(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	CreateWatchlist(db, "Tech")
	CreateWatchlist(db, "Crypto")

	lists, err := ListWatchlists(db)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(lists) != 2 {
		t.Errorf("expected 2 watchlists, got %d", len(lists))
	}
}

func TestListWatchlistItems(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	wID, _ := CreateWatchlist(db, "Mixed")
	AddWatchlistItem(db, wID, "AAPL", "us_equity")
	AddWatchlistItem(db, wID, "BTC", "crypto")

	items, err := ListWatchlistItems(db, wID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestRemoveWatchlistItem(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	wID, _ := CreateWatchlist(db, "Tech")
	AddWatchlistItem(db, wID, "AAPL", "us_equity")
	AddWatchlistItem(db, wID, "GOOG", "us_equity")

	if err := RemoveWatchlistItem(db, wID, "AAPL"); err != nil {
		t.Fatalf("remove: %v", err)
	}

	items, _ := ListWatchlistItems(db, wID)
	if len(items) != 1 {
		t.Errorf("expected 1 item after remove, got %d", len(items))
	}
	if items[0].Symbol != "GOOG" {
		t.Errorf("expected GOOG remaining, got %s", items[0].Symbol)
	}
}

func TestDeleteWatchlistCascade(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	wID, _ := CreateWatchlist(db, "Tech")
	AddWatchlistItem(db, wID, "AAPL", "us_equity")
	AddWatchlistItem(db, wID, "GOOG", "us_equity")

	if err := DeleteWatchlist(db, wID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	lists, _ := ListWatchlists(db)
	if len(lists) != 0 {
		t.Errorf("expected 0 watchlists after delete, got %d", len(lists))
	}

	// Items should be cascade-deleted
	items, _ := ListWatchlistItems(db, wID)
	if len(items) != 0 {
		t.Errorf("expected 0 items after cascade delete, got %d", len(items))
	}
}

func TestAllWatchedSymbols(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	w1, _ := CreateWatchlist(db, "List1")
	w2, _ := CreateWatchlist(db, "List2")
	AddWatchlistItem(db, w1, "AAPL", "us_equity")
	AddWatchlistItem(db, w1, "GOOG", "us_equity")
	AddWatchlistItem(db, w2, "AAPL", "us_equity") // duplicate across lists
	AddWatchlistItem(db, w2, "BTC", "crypto")

	symbols, err := AllWatchedSymbols(db)
	if err != nil {
		t.Fatalf("all watched: %v", err)
	}
	if len(symbols) != 3 {
		t.Errorf("expected 3 unique symbols, got %d: %v", len(symbols), symbols)
	}
}

func TestAddDuplicateWatchlistItem(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	wID, _ := CreateWatchlist(db, "Tech")
	AddWatchlistItem(db, wID, "AAPL", "us_equity")

	// Adding the same symbol again should not error (idempotent)
	err = AddWatchlistItem(db, wID, "AAPL", "us_equity")
	if err != nil {
		t.Errorf("duplicate add should be idempotent, got: %v", err)
	}

	items, _ := ListWatchlistItems(db, wID)
	if len(items) != 1 {
		t.Errorf("expected 1 item (no duplicate), got %d", len(items))
	}
}
