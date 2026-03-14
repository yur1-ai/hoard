# Known Issues & Technical Debt

Tracked items that are not critical but should be addressed.

## 1. Logger file handle leak (low priority)

**File:** `internal/logger/logger.go`

`logger.Init()` opens a file for slog but never exposes a way to close it. The handle leaks for the process lifetime. For a short-lived CLI/TUI this is fine, but it prevents clean shutdown and makes `Init()` unsafe to call more than once.

**Fix:** Return a `func() error` cleanup callback from `Init`, or expose a `Shutdown()` function. Wire it into `main.go` with a `defer`.

## 2. Missing database indexes (medium priority)

**File:** `migrations/001_initial_schema.sql`

No indexes on frequently queried columns:
- `holdings(account_id)` — used by `ListHoldings`
- `holdings(account_id, symbol, market)` — used by `applyBuy`/`applySell` on every transaction
- `transactions(account_id, symbol)` — used by `ListTransactions`

For a personal finance app the data volume is small, so this won't cause visible slowness until hundreds of holdings. Should be added in a future migration (002).

```sql
CREATE INDEX IF NOT EXISTS idx_holdings_account ON holdings(account_id);
CREATE INDEX IF NOT EXISTS idx_holdings_lookup ON holdings(account_id, symbol, market);
CREATE INDEX IF NOT EXISTS idx_transactions_symbol ON transactions(account_id, symbol);
```
