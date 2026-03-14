package market

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFinnhubGetQuote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/quote" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "AAPL" {
			t.Errorf("unexpected symbol: %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("token") != "test-key" {
			t.Errorf("unexpected token: %s", r.URL.Query().Get("token"))
		}
		json.NewEncoder(w).Encode(finnhubQuoteResponse{
			Current:    175.50,
			Change:     2.30,
			ChangePct:  1.33,
			High:       176.00,
			Low:        173.20,
			Open:       173.50,
			PrevClose:  173.20,
		})
	}))
	defer srv.Close()

	client := NewFinnhubClient("test-key", srv.Client())
	client.baseURL = srv.URL

	q, err := client.GetQuote(context.Background(), "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if q.Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", q.Symbol)
	}
	if q.Price != 175.50 {
		t.Errorf("expected 175.50, got %f", q.Price)
	}
	if q.Change != 2.30 {
		t.Errorf("expected 2.30, got %f", q.Change)
	}
	if q.ChangePct != 1.33 {
		t.Errorf("expected 1.33, got %f", q.ChangePct)
	}
}

func TestFinnhubGetQuotes(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		sym := r.URL.Query().Get("symbol")
		prices := map[string]float64{"AAPL": 175.0, "GOOG": 140.0, "MSFT": 415.0}
		json.NewEncoder(w).Encode(finnhubQuoteResponse{
			Current: prices[sym],
		})
	}))
	defer srv.Close()

	client := NewFinnhubClient("test-key", srv.Client())
	client.baseURL = srv.URL

	quotes, err := client.GetQuotes(context.Background(), []string{"AAPL", "GOOG", "MSFT"})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 3 {
		t.Fatalf("expected 3 quotes, got %d", len(quotes))
	}
	if callCount != 3 {
		t.Errorf("expected 3 API calls, got %d", callCount)
	}
}

func TestFinnhubSearchSymbol(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(finnhubSearchResponse{
			Result: []finnhubSearchResult{
				{Description: "Apple Inc", DisplaySymbol: "AAPL", Symbol: "AAPL", Type: "Common Stock"},
				{Description: "Apple Hospitality", DisplaySymbol: "APLE", Symbol: "APLE", Type: "Common Stock"},
			},
		})
	}))
	defer srv.Close()

	client := NewFinnhubClient("test-key", srv.Client())
	client.baseURL = srv.URL

	matches, err := client.SearchSymbol(context.Background(), "apple")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", matches[0].Symbol)
	}
	if matches[0].Description != "Apple Inc" {
		t.Errorf("expected 'Apple Inc', got %s", matches[0].Description)
	}
}

func TestFinnhubTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		json.NewEncoder(w).Encode(finnhubQuoteResponse{Current: 100})
	}))
	defer srv.Close()

	client := NewFinnhubClient("test-key", srv.Client())
	client.baseURL = srv.URL

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetQuote(ctx, "AAPL")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestFinnhubEmptySymbols(t *testing.T) {
	client := NewFinnhubClient("test-key", nil)
	quotes, err := client.GetQuotes(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 0 {
		t.Errorf("expected 0 quotes, got %d", len(quotes))
	}
}

func TestFinnhubZeroPriceSkipped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Finnhub returns zero values for invalid symbols
		json.NewEncoder(w).Encode(finnhubQuoteResponse{})
	}))
	defer srv.Close()

	client := NewFinnhubClient("test-key", srv.Client())
	client.baseURL = srv.URL

	q, err := client.GetQuote(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for zero-price quote")
	}
	if q != nil {
		t.Error("expected nil quote")
	}
}
