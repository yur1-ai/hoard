package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Account represents a financial account (brokerage, retirement, crypto wallet).
type Account struct {
	ID        int64
	Name      string
	Type      string
	Currency  string
	CreatedAt time.Time
}

// Holding represents a position in a single security or crypto asset.
type Holding struct {
	ID           int64
	AccountID    int64
	Symbol       string
	Market       string
	Quantity     float64
	AvgCostBasis float64
	Notes        string
}

// Transaction represents a single buy/sell/dividend/transfer event.
type Transaction struct {
	ID        int64
	AccountID int64
	Symbol    string
	Market    string
	Type      string
	Quantity  float64
	Price     float64
	Fee       float64
	Date      time.Time
	Notes     string
}

func CreateAccount(db *sql.DB, name, typ, currency string) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO accounts (name, type, currency) VALUES (?, ?, ?)",
		name, typ, currency,
	)
	if err != nil {
		return 0, fmt.Errorf("create account: %w", err)
	}
	return res.LastInsertId()
}

func ListAccounts(db *sql.DB) ([]Account, error) {
	rows, err := db.Query("SELECT id, name, type, currency, created_at FROM accounts ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Currency, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func CreateHolding(db *sql.DB, h Holding) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO holdings (account_id, symbol, market, quantity, avg_cost_basis, notes) VALUES (?, ?, ?, ?, ?, ?)",
		h.AccountID, h.Symbol, h.Market, h.Quantity, h.AvgCostBasis, h.Notes,
	)
	if err != nil {
		return 0, fmt.Errorf("create holding: %w", err)
	}
	return res.LastInsertId()
}

func ListHoldings(db *sql.DB, accountID int64) ([]Holding, error) {
	var query string
	var args []any
	if accountID == 0 {
		query = "SELECT id, account_id, symbol, market, quantity, avg_cost_basis, COALESCE(notes, '') FROM holdings ORDER BY symbol"
	} else {
		query = "SELECT id, account_id, symbol, market, quantity, avg_cost_basis, COALESCE(notes, '') FROM holdings WHERE account_id = ? ORDER BY symbol"
		args = append(args, accountID)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list holdings: %w", err)
	}
	defer rows.Close()

	var holdings []Holding
	for rows.Next() {
		var h Holding
		if err := rows.Scan(&h.ID, &h.AccountID, &h.Symbol, &h.Market, &h.Quantity, &h.AvgCostBasis, &h.Notes); err != nil {
			return nil, fmt.Errorf("scan holding: %w", err)
		}
		holdings = append(holdings, h)
	}
	return holdings, rows.Err()
}

func UpdateHolding(db *sql.DB, h Holding) error {
	_, err := db.Exec(
		"UPDATE holdings SET symbol = ?, market = ?, quantity = ?, avg_cost_basis = ?, notes = ? WHERE id = ?",
		h.Symbol, h.Market, h.Quantity, h.AvgCostBasis, h.Notes, h.ID,
	)
	if err != nil {
		return fmt.Errorf("update holding: %w", err)
	}
	return nil
}

func DeleteHolding(db *sql.DB, id int64) error {
	res, err := db.Exec("DELETE FROM holdings WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete holding: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("holding %d not found", id)
	}
	return nil
}

// AddTransaction records a transaction and updates the corresponding holding.
// For buy transactions, it auto-creates the holding if none exists and recalculates avg cost basis.
// For sell transactions, it reduces quantity (errors if insufficient).
// Dividend and transfer transactions are recorded without modifying holdings.
func AddTransaction(db *sql.DB, tx Transaction) (int64, error) {
	dbTx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback()

	res, err := dbTx.Exec(
		"INSERT INTO transactions (account_id, symbol, market, type, quantity, price, fee, date, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		tx.AccountID, tx.Symbol, tx.Market, tx.Type, tx.Quantity, tx.Price, tx.Fee, tx.Date, tx.Notes,
	)
	if err != nil {
		return 0, fmt.Errorf("insert transaction: %w", err)
	}
	txID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	switch tx.Type {
	case "buy":
		if err := applyBuy(dbTx, tx); err != nil {
			return 0, err
		}
	case "sell":
		if err := applySell(dbTx, tx); err != nil {
			return 0, err
		}
	case "dividend", "transfer":
		// Recorded only — no holding modification
	}

	if err := dbTx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return txID, nil
}

func applyBuy(dbTx *sql.Tx, tx Transaction) error {
	if tx.Quantity <= 0 {
		return fmt.Errorf("buy quantity must be positive, got %f", tx.Quantity)
	}

	var holdingID int64
	var oldQty, oldAvg float64

	err := dbTx.QueryRow(
		"SELECT id, quantity, avg_cost_basis FROM holdings WHERE account_id = ? AND symbol = ? AND market = ?",
		tx.AccountID, tx.Symbol, tx.Market,
	).Scan(&holdingID, &oldQty, &oldAvg)

	if err == sql.ErrNoRows {
		// First buy — create the holding
		_, err = dbTx.Exec(
			"INSERT INTO holdings (account_id, symbol, market, quantity, avg_cost_basis) VALUES (?, ?, ?, ?, ?)",
			tx.AccountID, tx.Symbol, tx.Market, tx.Quantity, tx.Price,
		)
		if err != nil {
			return fmt.Errorf("create holding on buy: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("lookup holding: %w", err)
	}

	// Update existing holding with weighted average cost basis
	newQty := oldQty + tx.Quantity
	newAvg := ((oldQty * oldAvg) + (tx.Quantity * tx.Price)) / newQty

	_, err = dbTx.Exec(
		"UPDATE holdings SET quantity = ?, avg_cost_basis = ? WHERE id = ?",
		newQty, newAvg, holdingID,
	)
	if err != nil {
		return fmt.Errorf("update holding on buy: %w", err)
	}
	return nil
}

func applySell(dbTx *sql.Tx, tx Transaction) error {
	if tx.Quantity <= 0 {
		return fmt.Errorf("sell quantity must be positive, got %f", tx.Quantity)
	}

	var holdingID int64
	var oldQty float64

	err := dbTx.QueryRow(
		"SELECT id, quantity FROM holdings WHERE account_id = ? AND symbol = ? AND market = ?",
		tx.AccountID, tx.Symbol, tx.Market,
	).Scan(&holdingID, &oldQty)

	if err == sql.ErrNoRows {
		return fmt.Errorf("no holding for %s: cannot sell", tx.Symbol)
	}
	if err != nil {
		return fmt.Errorf("lookup holding for sell: %w", err)
	}

	if tx.Quantity > oldQty {
		return fmt.Errorf("insufficient quantity: have %.4f, selling %.4f", oldQty, tx.Quantity)
	}

	newQty := oldQty - tx.Quantity
	_, err = dbTx.Exec("UPDATE holdings SET quantity = ? WHERE id = ?", newQty, holdingID)
	if err != nil {
		return fmt.Errorf("update holding on sell: %w", err)
	}
	return nil
}

func ListTransactions(db *sql.DB, symbol string) ([]Transaction, error) {
	rows, err := db.Query(
		"SELECT id, account_id, symbol, market, type, quantity, price, fee, date, COALESCE(notes, '') FROM transactions WHERE symbol = ? ORDER BY date DESC",
		symbol,
	)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.AccountID, &tx.Symbol, &tx.Market, &tx.Type, &tx.Quantity, &tx.Price, &tx.Fee, &tx.Date, &tx.Notes); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}

func AllEquitySymbols(db *sql.DB) ([]string, error) {
	return symbolsByMarket(db, "us_equity")
}

func AllCryptoSymbols(db *sql.DB) ([]string, error) {
	return symbolsByMarket(db, "crypto")
}

func symbolsByMarket(db *sql.DB, market string) ([]string, error) {
	rows, err := db.Query(
		"SELECT DISTINCT symbol FROM holdings WHERE market = ? ORDER BY symbol",
		market,
	)
	if err != nil {
		return nil, fmt.Errorf("symbols by market %s: %w", market, err)
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
