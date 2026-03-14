package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type CachedQuote struct {
	Symbol      string
	Market      string
	Price       float64
	Change      float64
	ChangePct   float64
	Volume      float64
	High24h     float64
	Low24h      float64
	LastUpdated time.Time
}

func SetQuote(db *sql.DB, q CachedQuote) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO market_cache
			(symbol, market, price, change, change_pct, volume, high_24h, low_24h, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		q.Symbol, q.Market, q.Price, q.Change, q.ChangePct, q.Volume, q.High24h, q.Low24h, q.LastUpdated,
	)
	if err != nil {
		return fmt.Errorf("set quote: %w", err)
	}
	return nil
}

// SetQuotes bulk-upserts quotes in a single transaction.
func SetQuotes(db *sql.DB, quotes []CachedQuote) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO market_cache
			(symbol, market, price, change, change_pct, volume, high_24h, low_24h, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, q := range quotes {
		if _, err := stmt.Exec(q.Symbol, q.Market, q.Price, q.Change, q.ChangePct, q.Volume, q.High24h, q.Low24h, q.LastUpdated); err != nil {
			return fmt.Errorf("exec %s: %w", q.Symbol, err)
		}
	}

	return tx.Commit()
}

func GetQuote(db *sql.DB, symbol string) (*CachedQuote, error) {
	var q CachedQuote
	err := db.QueryRow(
		"SELECT symbol, market, COALESCE(price, 0), COALESCE(change, 0), COALESCE(change_pct, 0), COALESCE(volume, 0), COALESCE(high_24h, 0), COALESCE(low_24h, 0), last_updated FROM market_cache WHERE symbol = ?",
		symbol,
	).Scan(&q.Symbol, &q.Market, &q.Price, &q.Change, &q.ChangePct, &q.Volume, &q.High24h, &q.Low24h, &q.LastUpdated)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get quote: %w", err)
	}
	return &q, nil
}

func GetQuotes(db *sql.DB, symbols []string) ([]CachedQuote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	placeholders := strings.Repeat("?,", len(symbols))
	placeholders = placeholders[:len(placeholders)-1] // trim trailing comma

	args := make([]any, len(symbols))
	for i, s := range symbols {
		args[i] = s
	}

	rows, err := db.Query(
		fmt.Sprintf("SELECT symbol, market, COALESCE(price, 0), COALESCE(change, 0), COALESCE(change_pct, 0), COALESCE(volume, 0), COALESCE(high_24h, 0), COALESCE(low_24h, 0), last_updated FROM market_cache WHERE symbol IN (%s)", placeholders),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("get quotes: %w", err)
	}
	defer rows.Close()

	var quotes []CachedQuote
	for rows.Next() {
		var q CachedQuote
		if err := rows.Scan(&q.Symbol, &q.Market, &q.Price, &q.Change, &q.ChangePct, &q.Volume, &q.High24h, &q.Low24h, &q.LastUpdated); err != nil {
			return nil, fmt.Errorf("scan quote: %w", err)
		}
		quotes = append(quotes, q)
	}
	return quotes, rows.Err()
}

// IsStale returns true if the quote is older than maxAge.
func IsStale(q *CachedQuote, maxAge time.Duration) bool {
	if q == nil {
		return true
	}
	return time.Since(q.LastUpdated) > maxAge
}

func SetCurrencyRate(db *sql.DB, from, to string, rate float64) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO currency_rates (from_currency, to_currency, rate, fetched_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		from, to, rate,
	)
	if err != nil {
		return fmt.Errorf("set currency rate: %w", err)
	}
	return nil
}

func GetCurrencyRate(db *sql.DB, from, to string) (float64, error) {
	var rate float64
	err := db.QueryRow(
		"SELECT rate FROM currency_rates WHERE from_currency = ? AND to_currency = ?",
		from, to,
	).Scan(&rate)
	if err != nil {
		return 0, fmt.Errorf("get currency rate %s→%s: %w", from, to, err)
	}
	return rate, nil
}
