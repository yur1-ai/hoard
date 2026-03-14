package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const frankfurterDefaultURL = "https://api.frankfurter.dev"

// FrankfurterClient fetches currency exchange rates (no API key needed).
type FrankfurterClient struct {
	client  *http.Client
	baseURL string
}

func NewFrankfurterClient(httpClient *http.Client) *FrankfurterClient {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &FrankfurterClient{
		client:  httpClient,
		baseURL: frankfurterDefaultURL,
	}
}

type frankfurterResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

// FetchRates returns exchange rates from the given base currency to all available targets.
func (c *FrankfurterClient) FetchRates(ctx context.Context, base string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/v1/latest?base=%s", c.baseURL, base)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("frankfurter: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("frankfurter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("frankfurter: status %d", resp.StatusCode)
	}

	var data frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("frankfurter decode: %w", err)
	}

	return data.Rates, nil
}
