package market

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

const finnhubDefaultURL = "https://finnhub.io"

// FinnhubClient implements StockProvider using the Finnhub REST API.
type FinnhubClient struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

func NewFinnhubClient(apiKey string, httpClient *http.Client) *FinnhubClient {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &FinnhubClient{
		apiKey:  apiKey,
		client:  httpClient,
		baseURL: finnhubDefaultURL,
	}
}

// finnhubQuoteResponse maps Finnhub /api/v1/quote JSON.
type finnhubQuoteResponse struct {
	Current   float64 `json:"c"`
	Change    float64 `json:"d"`
	ChangePct float64 `json:"dp"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Open      float64 `json:"o"`
	PrevClose float64 `json:"pc"`
}

type finnhubSearchResponse struct {
	Result []finnhubSearchResult `json:"result"`
}

type finnhubSearchResult struct {
	Description   string `json:"description"`
	DisplaySymbol string `json:"displaySymbol"`
	Symbol        string `json:"symbol"`
	Type          string `json:"type"`
}

func (c *FinnhubClient) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	u := fmt.Sprintf("%s/api/v1/quote?symbol=%s&token=%s", c.baseURL, url.QueryEscape(symbol), url.QueryEscape(c.apiKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("finnhub quote %s: %w", symbol, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("finnhub quote %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("finnhub quote %s: status %d", symbol, resp.StatusCode)
	}

	var data finnhubQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("finnhub quote %s decode: %w", symbol, err)
	}

	// Finnhub returns all zeros for invalid symbols
	if data.Current == 0 {
		return nil, fmt.Errorf("finnhub quote %s: no data (invalid symbol or market closed)", symbol)
	}

	return &Quote{
		Symbol:    symbol,
		Price:     data.Current,
		Change:    data.Change,
		ChangePct: data.ChangePct,
		High:      data.High,
		Low:       data.Low,
	}, nil
}

func (c *FinnhubClient) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	var quotes []Quote
	var lastErr error
	for _, sym := range symbols {
		q, err := c.GetQuote(ctx, sym)
		if err != nil {
			slog.Warn("failed to fetch quote", "symbol", sym, "error", err)
			lastErr = err
			continue
		}
		quotes = append(quotes, *q)
	}
	// If ALL symbols failed, return the error so callers know something is wrong
	if len(quotes) == 0 && lastErr != nil {
		return nil, fmt.Errorf("all %d symbol lookups failed, last: %w", len(symbols), lastErr)
	}
	return quotes, nil
}

func (c *FinnhubClient) SearchSymbol(ctx context.Context, query string) ([]SymbolMatch, error) {
	u := fmt.Sprintf("%s/api/v1/search?q=%s&token=%s", c.baseURL, url.QueryEscape(query), url.QueryEscape(c.apiKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("finnhub search: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("finnhub search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("finnhub search: status %d", resp.StatusCode)
	}

	var data finnhubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("finnhub search decode: %w", err)
	}

	matches := make([]SymbolMatch, len(data.Result))
	for i, r := range data.Result {
		matches[i] = SymbolMatch{
			Symbol:      r.Symbol,
			Description: r.Description,
			Type:        r.Type,
		}
	}
	return matches, nil
}

// Compile-time check that FinnhubClient implements StockProvider.
var _ StockProvider = (*FinnhubClient)(nil)
