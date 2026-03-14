package currency

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFrankfurterFetchRates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("base") != "USD" {
			t.Errorf("unexpected base: %s", r.URL.Query().Get("base"))
		}
		json.NewEncoder(w).Encode(frankfurterResponse{
			Base:  "USD",
			Date:  "2026-03-14",
			Rates: map[string]float64{"EUR": 0.92, "GBP": 0.79, "JPY": 149.50},
		})
	}))
	defer srv.Close()

	client := NewFrankfurterClient(srv.Client())
	client.baseURL = srv.URL

	rates, err := client.FetchRates(context.Background(), "USD")
	if err != nil {
		t.Fatal(err)
	}
	if len(rates) != 3 {
		t.Fatalf("expected 3 rates, got %d", len(rates))
	}
	if rates["EUR"] != 0.92 {
		t.Errorf("expected EUR=0.92, got %f", rates["EUR"])
	}
	if rates["GBP"] != 0.79 {
		t.Errorf("expected GBP=0.79, got %f", rates["GBP"])
	}
}

func TestFrankfurterTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		json.NewEncoder(w).Encode(frankfurterResponse{})
	}))
	defer srv.Close()

	client := NewFrankfurterClient(srv.Client())
	client.baseURL = srv.URL

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.FetchRates(ctx, "USD")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestFrankfurterBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewFrankfurterClient(srv.Client())
	client.baseURL = srv.URL

	_, err := client.FetchRates(context.Background(), "USD")
	if err == nil {
		t.Fatal("expected error for 429 status")
	}
}
