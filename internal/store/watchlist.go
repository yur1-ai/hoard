package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Watchlist struct {
	ID   int64
	Name string
}

type WatchlistItem struct {
	WatchlistID int64
	Symbol      string
	Market      string
	AddedAt     time.Time
}

func CreateWatchlist(db *sql.DB, name string) (int64, error) {
	res, err := db.Exec("INSERT INTO watchlists (name) VALUES (?)", name)
	if err != nil {
		return 0, fmt.Errorf("create watchlist: %w", err)
	}
	return res.LastInsertId()
}

func ListWatchlists(db *sql.DB) ([]Watchlist, error) {
	rows, err := db.Query("SELECT id, name FROM watchlists ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list watchlists: %w", err)
	}
	defer rows.Close()

	var lists []Watchlist
	for rows.Next() {
		var w Watchlist
		if err := rows.Scan(&w.ID, &w.Name); err != nil {
			return nil, fmt.Errorf("scan watchlist: %w", err)
		}
		lists = append(lists, w)
	}
	return lists, rows.Err()
}

func DeleteWatchlist(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM watchlists WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete watchlist: %w", err)
	}
	return nil
}

// AddWatchlistItem adds a symbol to a watchlist. Idempotent — duplicates are ignored.
func AddWatchlistItem(db *sql.DB, watchlistID int64, symbol, market string) error {
	_, err := db.Exec(
		"INSERT OR IGNORE INTO watchlist_items (watchlist_id, symbol, market) VALUES (?, ?, ?)",
		watchlistID, symbol, market,
	)
	if err != nil {
		return fmt.Errorf("add watchlist item: %w", err)
	}
	return nil
}

func RemoveWatchlistItem(db *sql.DB, watchlistID int64, symbol string) error {
	_, err := db.Exec(
		"DELETE FROM watchlist_items WHERE watchlist_id = ? AND symbol = ?",
		watchlistID, symbol,
	)
	if err != nil {
		return fmt.Errorf("remove watchlist item: %w", err)
	}
	return nil
}

func ListWatchlistItems(db *sql.DB, watchlistID int64) ([]WatchlistItem, error) {
	rows, err := db.Query(
		"SELECT watchlist_id, symbol, market, added_at FROM watchlist_items WHERE watchlist_id = ? ORDER BY symbol",
		watchlistID,
	)
	if err != nil {
		return nil, fmt.Errorf("list watchlist items: %w", err)
	}
	defer rows.Close()

	var items []WatchlistItem
	for rows.Next() {
		var item WatchlistItem
		if err := rows.Scan(&item.WatchlistID, &item.Symbol, &item.Market, &item.AddedAt); err != nil {
			return nil, fmt.Errorf("scan watchlist item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// AllWatchedSymbols returns unique symbols across all watchlists.
func AllWatchedSymbols(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT symbol FROM watchlist_items ORDER BY symbol")
	if err != nil {
		return nil, fmt.Errorf("all watched symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		symbols = append(symbols, s)
	}
	return symbols, rows.Err()
}
