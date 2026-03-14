package market

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/yur1-ai/hoard/internal/store"
)

// mockStock implements StockProvider for testing.
type mockStock struct {
	quotes map[string]*Quote
	err    error
	calls  int
}

func (m *mockStock) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	q, ok := m.quotes[symbol]
	if !ok {
		return nil, nil
	}
	return q, nil
}

func (m *mockStock) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	var out []Quote
	for _, s := range symbols {
		if q, ok := m.quotes[s]; ok {
			out = append(out, *q)
		}
	}
	return out, nil
}

func (m *mockStock) SearchSymbol(ctx context.Context, query string) ([]SymbolMatch, error) {
	return nil, nil
}

// mockCrypto implements CryptoProvider for testing.
type mockCrypto struct {
	quotes map[string]*Quote
	err    error
	calls  int
}

func (m *mockCrypto) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	q, ok := m.quotes[symbol]
	if !ok {
		return nil, nil
	}
	return q, nil
}

func (m *mockCrypto) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	var out []Quote
	for _, s := range symbols {
		if q, ok := m.quotes[s]; ok {
			out = append(out, *q)
		}
	}
	return out, nil
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCacheServiceFetchesFromProvider(t *testing.T) {
	db := openTestDB(t)
	stocks := &mockStock{
		quotes: map[string]*Quote{
			"AAPL": {Symbol: "AAPL", Price: 175.0, Change: 2.0, ChangePct: 1.1},
		},
	}

	svc := NewCachedService(stocks, nil, db, 30*time.Second, 120*time.Second)

	quotes, err := svc.GetStockQuotes(context.Background(), []string{"AAPL"})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 {
		t.Fatalf("expected 1 quote, got %d", len(quotes))
	}
	if quotes[0].Price != 175.0 {
		t.Errorf("expected 175.0, got %f", quotes[0].Price)
	}
	if stocks.calls != 1 {
		t.Errorf("expected 1 provider call, got %d", stocks.calls)
	}
}

func TestCacheServiceReturnsCached(t *testing.T) {
	db := openTestDB(t)

	// Pre-populate cache
	store.SetQuote(db, store.CachedQuote{
		Symbol: "AAPL", Market: "equity", Price: 170.0,
		LastUpdated: time.Now(), // fresh
	})

	stocks := &mockStock{
		quotes: map[string]*Quote{
			"AAPL": {Symbol: "AAPL", Price: 175.0},
		},
	}

	svc := NewCachedService(stocks, nil, db, 30*time.Second, 120*time.Second)

	quotes, err := svc.GetStockQuotes(context.Background(), []string{"AAPL"})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 {
		t.Fatalf("expected 1 quote, got %d", len(quotes))
	}
	// Should return cached price, not provider price
	if quotes[0].Price != 170.0 {
		t.Errorf("expected cached 170.0, got %f", quotes[0].Price)
	}
	// Provider should NOT have been called
	if stocks.calls != 0 {
		t.Errorf("expected 0 provider calls, got %d", stocks.calls)
	}
}

func TestCacheServiceStaleFallback(t *testing.T) {
	db := openTestDB(t)

	// Pre-populate with stale cache (2 minutes old, TTL is 30s)
	store.SetQuote(db, store.CachedQuote{
		Symbol: "AAPL", Market: "equity", Price: 170.0,
		LastUpdated: time.Now().Add(-2 * time.Minute),
	})

	stocks := &mockStock{
		err: context.DeadlineExceeded, // provider fails
	}

	svc := NewCachedService(stocks, nil, db, 30*time.Second, 120*time.Second)

	quotes, err := svc.GetStockQuotes(context.Background(), []string{"AAPL"})
	// Should succeed with stale data, not error
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 {
		t.Fatalf("expected 1 stale quote, got %d", len(quotes))
	}
	if quotes[0].Price != 170.0 {
		t.Errorf("expected stale 170.0, got %f", quotes[0].Price)
	}
}

func TestCacheServiceCryptoQuotes(t *testing.T) {
	db := openTestDB(t)
	crypto := &mockCrypto{
		quotes: map[string]*Quote{
			"BTC-USD": {Symbol: "BTC-USD", Price: 67500.0},
		},
	}

	svc := NewCachedService(nil, crypto, db, 30*time.Second, 120*time.Second)

	quotes, err := svc.GetCryptoQuotes(context.Background(), []string{"BTC-USD"})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 {
		t.Fatalf("expected 1 quote, got %d", len(quotes))
	}
	if quotes[0].Price != 67500.0 {
		t.Errorf("expected 67500, got %f", quotes[0].Price)
	}
}
