package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const coingeckoDefaultURL = "https://api.coingecko.com"

// CoinGeckoClient implements CryptoProvider using the CoinGecko REST API.
type CoinGeckoClient struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

func NewCoinGeckoClient(apiKey string, httpClient *http.Client) *CoinGeckoClient {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &CoinGeckoClient{
		apiKey:  apiKey,
		client:  httpClient,
		baseURL: coingeckoDefaultURL,
	}
}

// cgPriceData maps the CoinGecko /simple/price JSON for a single coin.
type cgPriceData struct {
	USD          float64 `json:"usd"`
	USD24hChange float64 `json:"usd_24h_change"`
	USD24hVol    float64 `json:"usd_24h_vol"`
	USDHigh24h   float64 `json:"usd_high_24h"`
	USDLow24h    float64 `json:"usd_low_24h"`
}

func (c *CoinGeckoClient) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	coinID, ok := resolveCoinID(symbol)
	if !ok {
		return nil, fmt.Errorf("coingecko: unknown symbol %q", symbol)
	}

	data, err := c.fetchPrices(ctx, []string{coinID})
	if err != nil {
		return nil, err
	}

	price, ok := data[coinID]
	if !ok {
		return nil, fmt.Errorf("coingecko: no data for %s (%s)", symbol, coinID)
	}

	return &Quote{
		Symbol:    symbol,
		Price:     price.USD,
		ChangePct: price.USD24hChange,
		Volume:    price.USD24hVol,
		High:      price.USDHigh24h,
		Low:       price.USDLow24h,
	}, nil
}

func (c *CoinGeckoClient) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	// Map symbols to CoinGecko IDs
	idToSymbol := make(map[string]string, len(symbols))
	var ids []string
	for _, sym := range symbols {
		coinID, ok := resolveCoinID(sym)
		if !ok {
			continue // skip unknown symbols
		}
		idToSymbol[coinID] = sym
		ids = append(ids, coinID)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	data, err := c.fetchPrices(ctx, ids)
	if err != nil {
		return nil, err
	}

	var quotes []Quote
	for coinID, price := range data {
		sym, ok := idToSymbol[coinID]
		if !ok {
			continue
		}
		quotes = append(quotes, Quote{
			Symbol:    sym,
			Price:     price.USD,
			ChangePct: price.USD24hChange,
			Volume:    price.USD24hVol,
			High:      price.USDHigh24h,
			Low:       price.USDLow24h,
		})
	}
	return quotes, nil
}

func (c *CoinGeckoClient) fetchPrices(ctx context.Context, ids []string) (map[string]cgPriceData, error) {
	url := fmt.Sprintf(
		"%s/api/v3/simple/price?ids=%s&vs_currencies=usd&include_24hr_change=true&include_24hr_vol=true&include_high_24h=true&include_low_24h=true",
		c.baseURL, strings.Join(ids, ","),
	)
	if c.apiKey != "" {
		url += "&x_cg_demo_api_key=" + c.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("coingecko: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coingecko: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko: status %d", resp.StatusCode)
	}

	var data map[string]cgPriceData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("coingecko decode: %w", err)
	}
	return data, nil
}

var _ CryptoProvider = (*CoinGeckoClient)(nil)
