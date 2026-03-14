package store

import (
	"testing"
	"time"
)

func TestSetAndGetQuote(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	q := CachedQuote{
		Symbol:      "AAPL",
		Market:      "us_equity",
		Price:       175.50,
		Change:      2.30,
		ChangePct:   1.33,
		Volume:      45000000,
		High24h:     176.00,
		Low24h:      173.00,
		LastUpdated: time.Now(),
	}

	if err := SetQuote(db, q); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, err := GetQuote(db, "AAPL")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected quote, got nil")
	}
	if got.Price != 175.50 {
		t.Errorf("expected price 175.50, got %f", got.Price)
	}
	if got.Market != "us_equity" {
		t.Errorf("expected market us_equity, got %s", got.Market)
	}
}

func TestGetQuoteNotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	got, err := GetQuote(db, "NONEXISTENT")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Error("expected nil for missing quote")
	}
}

func TestGetQuoteStale(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	fresh := CachedQuote{
		Symbol:      "AAPL",
		Market:      "us_equity",
		Price:       175.50,
		LastUpdated: time.Now(),
	}
	stale := CachedQuote{
		Symbol:      "GOOG",
		Market:      "us_equity",
		Price:       140.00,
		LastUpdated: time.Now().Add(-10 * time.Minute),
	}
	SetQuote(db, fresh)
	SetQuote(db, stale)

	maxAge := 5 * time.Minute

	freshQ, _ := GetQuote(db, "AAPL")
	if IsStale(freshQ, maxAge) {
		t.Error("AAPL should not be stale")
	}

	staleQ, _ := GetQuote(db, "GOOG")
	if !IsStale(staleQ, maxAge) {
		t.Error("GOOG should be stale")
	}
}

func TestBulkSetQuotes(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	now := time.Now()
	quotes := []CachedQuote{
		{Symbol: "AAPL", Market: "us_equity", Price: 175.50, LastUpdated: now},
		{Symbol: "GOOG", Market: "us_equity", Price: 140.00, LastUpdated: now},
		{Symbol: "BTC", Market: "crypto", Price: 60000, LastUpdated: now},
	}

	if err := SetQuotes(db, quotes); err != nil {
		t.Fatalf("bulk set: %v", err)
	}

	got, err := GetQuotes(db, []string{"AAPL", "BTC"})
	if err != nil {
		t.Fatalf("get quotes: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 quotes, got %d", len(got))
	}
}

func TestBulkSetQuotesOverwrite(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	now := time.Now()
	SetQuote(db, CachedQuote{Symbol: "AAPL", Market: "us_equity", Price: 100, LastUpdated: now})

	// Overwrite with new price
	SetQuotes(db, []CachedQuote{
		{Symbol: "AAPL", Market: "us_equity", Price: 200, LastUpdated: now},
	})

	got, _ := GetQuote(db, "AAPL")
	if got.Price != 200 {
		t.Errorf("expected updated price 200, got %f", got.Price)
	}
}

func TestSetAndGetCurrencyRate(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := SetCurrencyRate(db, "USD", "EUR", 0.92); err != nil {
		t.Fatalf("set rate: %v", err)
	}

	rate, err := GetCurrencyRate(db, "USD", "EUR")
	if err != nil {
		t.Fatalf("get rate: %v", err)
	}
	if rate != 0.92 {
		t.Errorf("expected 0.92, got %f", rate)
	}
}

func TestGetCurrencyRateNotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	_, err = GetCurrencyRate(db, "USD", "JPY")
	if err == nil {
		t.Error("expected error for missing rate")
	}
}

func TestSetCurrencyRateOverwrite(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	SetCurrencyRate(db, "USD", "EUR", 0.90)
	SetCurrencyRate(db, "USD", "EUR", 0.95)

	rate, _ := GetCurrencyRate(db, "USD", "EUR")
	if rate != 0.95 {
		t.Errorf("expected updated rate 0.95, got %f", rate)
	}
}
