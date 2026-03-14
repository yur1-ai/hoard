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

## 3. WindowSizeMsg not forwarded to sub-models (Phase 4)

**File:** `internal/app/update.go:10-15`

Currently `propagateSize()` calls `SetSize()` on sub-models but doesn't forward `tea.WindowSizeMsg` itself. This works with placeholder stubs but will need fixing in Phase 4 when sub-models have real `Init()`/`Update()` logic that reacts to resize events.

**Fix:** After `propagateSize()`, forward the message:
```go
m.sidebar, _ = m.sidebar.Update(msg)
m.portfolio, _ = m.portfolio.Update(msg)
// etc.
```

## 4. Header bar uses inline style instead of shared HeaderStyle (low priority)

**File:** `internal/ui/header/header.go:55-60`

The outer bar background is built with `lipgloss.NewStyle()` inline, duplicating the color from `common.ColorHeader`. Changes to the header color in `styles.go` won't propagate to the bar background.

## 5. Header formatNumber lacks thousands separator for mid-range values (Phase 4)

**File:** `internal/ui/header/header.go:64-72`

Values 1,000–999,999 render as `12346` without thousands separators or decimals. Should display as `12,345.67` or similar for a finance app. Fix when header gets real portfolio data in Phase 4.
