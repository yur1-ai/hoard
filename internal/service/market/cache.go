package market

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/yur1-ai/hoard/internal/store"
)

// CachedService wraps stock and crypto providers with SQLite-backed caching.
// On provider failure, returns stale cache data if available.
type CachedService struct {
	stocks   StockProvider
	crypto   CryptoProvider
	db       *sql.DB
	stockTTL time.Duration
	cryptoTTL time.Duration
}

func NewCachedService(stocks StockProvider, crypto CryptoProvider, db *sql.DB, stockTTL, cryptoTTL time.Duration) *CachedService {
	return &CachedService{
		stocks:    stocks,
		crypto:    crypto,
		db:        db,
		stockTTL:  stockTTL,
		cryptoTTL: cryptoTTL,
	}
}

// GetStockQuotes returns quotes for equity symbols, using cache when fresh.
func (s *CachedService) GetStockQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	return s.getQuotes(ctx, symbols, "equity", s.stockTTL, func(ctx context.Context, syms []string) ([]Quote, error) {
		if s.stocks == nil {
			return nil, nil
		}
		return s.stocks.GetQuotes(ctx, syms)
	})
}

// GetCryptoQuotes returns quotes for crypto symbols, using cache when fresh.
func (s *CachedService) GetCryptoQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	return s.getQuotes(ctx, symbols, "crypto", s.cryptoTTL, func(ctx context.Context, syms []string) ([]Quote, error) {
		if s.crypto == nil {
			return nil, nil
		}
		return s.crypto.GetQuotes(ctx, syms)
	})
}

// SearchSymbol delegates to the stock provider.
func (s *CachedService) SearchSymbol(ctx context.Context, query string) ([]SymbolMatch, error) {
	if s.stocks == nil {
		return nil, nil
	}
	return s.stocks.SearchSymbol(ctx, query)
}

func (s *CachedService) getQuotes(
	ctx context.Context,
	symbols []string,
	marketType string,
	ttl time.Duration,
	fetchFn func(context.Context, []string) ([]Quote, error),
) ([]Quote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	// Check cache for all symbols
	cached, err := store.GetQuotes(s.db, symbols)
	if err != nil {
		slog.Warn("cache read failed", "error", err)
	}

	cachedMap := make(map[string]store.CachedQuote, len(cached))
	for _, q := range cached {
		cachedMap[q.Symbol] = q
	}

	// Find stale or missing symbols
	var staleSymbols []string
	for _, sym := range symbols {
		cq, ok := cachedMap[sym]
		if !ok || store.IsStale(&cq, ttl) {
			staleSymbols = append(staleSymbols, sym)
		}
	}

	// If all cached and fresh, return immediately
	if len(staleSymbols) == 0 {
		return cachedToQuotes(cached), nil
	}

	// Fetch stale symbols from provider
	fresh, fetchErr := fetchFn(ctx, staleSymbols)
	if fetchErr != nil {
		slog.Warn("provider fetch failed, using stale cache", "error", fetchErr, "symbols", staleSymbols)
		// Fall back to whatever we have in cache (even stale)
		return cachedToQuotes(cached), nil
	}

	// Update cache with fresh data
	if len(fresh) > 0 {
		var cqs []store.CachedQuote
		for _, q := range fresh {
			cqs = append(cqs, store.CachedQuote{
				Symbol:      q.Symbol,
				Market:      marketType,
				Price:       q.Price,
				Change:      q.Change,
				ChangePct:   q.ChangePct,
				Volume:      q.Volume,
				High24h:     q.High,
				Low24h:      q.Low,
				LastUpdated: time.Now(),
			})
		}
		if err := store.SetQuotes(s.db, cqs); err != nil {
			slog.Warn("cache write failed", "error", err)
		}
	}

	// Merge: fresh data overrides cached
	result := make(map[string]Quote, len(symbols))
	for _, cq := range cached {
		result[cq.Symbol] = Quote{
			Symbol:    cq.Symbol,
			Price:     cq.Price,
			Change:    cq.Change,
			ChangePct: cq.ChangePct,
			Volume:    cq.Volume,
			High:      cq.High24h,
			Low:       cq.Low24h,
		}
	}
	for _, q := range fresh {
		result[q.Symbol] = q
	}

	var out []Quote
	for _, sym := range symbols {
		if q, ok := result[sym]; ok {
			out = append(out, q)
		}
	}
	return out, nil
}

func cachedToQuotes(cached []store.CachedQuote) []Quote {
	quotes := make([]Quote, len(cached))
	for i, cq := range cached {
		quotes[i] = Quote{
			Symbol:    cq.Symbol,
			Price:     cq.Price,
			Change:    cq.Change,
			ChangePct: cq.ChangePct,
			Volume:    cq.Volume,
			High:      cq.High24h,
			Low:       cq.Low24h,
		}
	}
	return quotes
}
