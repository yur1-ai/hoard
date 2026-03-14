package market

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCoinGeckoGetQuote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/simple/price" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		ids := r.URL.Query().Get("ids")
		if ids != "bitcoin" {
			t.Errorf("unexpected ids: %s", ids)
		}
		json.NewEncoder(w).Encode(map[string]cgPriceData{
			"bitcoin": {
				USD:          67500.00,
				USD24hChange: 2.5,
				USD24hVol:    35000000000,
				USDHigh24h:   68000.00,
				USDLow24h:    66500.00,
			},
		})
	}))
	defer srv.Close()

	client := NewCoinGeckoClient("test-key", srv.Client())
	client.baseURL = srv.URL

	q, err := client.GetQuote(context.Background(), "BTC-USD")
	if err != nil {
		t.Fatal(err)
	}
	if q.Symbol != "BTC-USD" {
		t.Errorf("expected BTC-USD, got %s", q.Symbol)
	}
	if q.Price != 67500.00 {
		t.Errorf("expected 67500, got %f", q.Price)
	}
	if q.ChangePct != 2.5 {
		t.Errorf("expected 2.5, got %f", q.ChangePct)
	}
}

func TestCoinGeckoGetQuotes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]cgPriceData{
			"bitcoin":  {USD: 67500},
			"ethereum": {USD: 3500},
		})
	}))
	defer srv.Close()

	client := NewCoinGeckoClient("test-key", srv.Client())
	client.baseURL = srv.URL

	quotes, err := client.GetQuotes(context.Background(), []string{"BTC-USD", "ETH-USD"})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}
}

func TestCoinGeckoUnknownSymbol(t *testing.T) {
	client := NewCoinGeckoClient("test-key", nil)
	_, err := client.GetQuote(context.Background(), "NOTACOIN-USD")
	if err == nil {
		t.Fatal("expected error for unknown symbol")
	}
}

func TestCoinGeckoTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]cgPriceData{"bitcoin": {USD: 67500}})
	}))
	defer srv.Close()

	client := NewCoinGeckoClient("test-key", srv.Client())
	client.baseURL = srv.URL

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetQuote(ctx, "BTC-USD")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestCoinGeckoEmptySymbols(t *testing.T) {
	client := NewCoinGeckoClient("test-key", nil)
	quotes, err := client.GetQuotes(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 0 {
		t.Errorf("expected 0 quotes, got %d", len(quotes))
	}
}

func TestResolveCoinID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		ok       bool
	}{
		{"BTC", "bitcoin", true},
		{"BTC-USD", "bitcoin", true},
		{"ETH-USD", "ethereum", true},
		{"eth", "ethereum", true},
		{"NOTACOIN", "", false},
	}
	for _, tt := range tests {
		id, ok := resolveCoinID(tt.input)
		if ok != tt.ok {
			t.Errorf("resolveCoinID(%q) ok=%v, want %v", tt.input, ok, tt.ok)
		}
		if id != tt.expected {
			t.Errorf("resolveCoinID(%q)=%q, want %q", tt.input, id, tt.expected)
		}
	}
}
