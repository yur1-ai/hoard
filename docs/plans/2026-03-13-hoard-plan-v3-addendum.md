# Hoard Implementation Plan — v3 Addendum

> This addendum patches both the original plan (`2026-03-13-hoard-implementation-plan.md`)
> and the v2 addendum (`2026-03-13-hoard-plan-v2-addendum.md`).
> It fixes 6 HIGH and 7 MEDIUM issues found during the final cross-document review.

---

## Reading Order

When implementing, apply patches in this order:
1. Read the original plan
2. Apply v2 addendum patches
3. Apply **this v3 addendum** — it supersedes conflicting content in both prior documents

---

## HIGH FIX 1: GoReleaser Config Duplicate — Task 8.4 SUPERSEDED

**Problem:** Phase 6.5 (v2 addendum, Task 6.5.2) creates a full `.goreleaser.yaml` with Homebrew/Scoop/AUR. Phase 8 (original, Task 8.4) also creates a simpler `.goreleaser.yaml`. An engineer would hit a conflict.

**Fix:** Task 8.4 in the original plan is **SUPERSEDED** by Task 6.5.2 in the v2 addendum. Skip Task 8.4 entirely. The `.goreleaser.yaml` is created once in Phase 6.5 and never recreated.

When you reach Phase 8, Task 8.4 should be checked off as "done — handled in Phase 6.5" with no action taken.

---

## HIGH FIX 2: Tab Key Sidebar State Machine

**Problem:** The code in original plan Task 1.5 (`model.go:909-914`) always sets `m.sidebarOpen = true`, meaning Tab can never close the sidebar.

**Supersedes:** Original plan Task 1.5, Step 2 — the `case "tab":` block in `App.Update()`.

**Fix:** Replace the Tab handler with a proper three-state machine:

```go
case "tab":
    if !m.sidebarOpen {
        // Sidebar closed → open it and focus it
        m.sidebarOpen = true
        m.focus = focusSidebar
    } else if m.focus == focusMarket {
        // Sidebar open, market focused → focus sidebar
        m.focus = focusSidebar
    } else {
        // Sidebar open, sidebar focused → close it and focus market
        m.sidebarOpen = false
        m.focus = focusMarket
    }
```

**State transitions:**

```
Tab pressed:                                Result:
─────────────────────────────────────────────────────────────────
Sidebar CLOSED, focus=market         →  Sidebar OPEN, focus=sidebar
Sidebar OPEN,   focus=market         →  Sidebar OPEN, focus=sidebar
Sidebar OPEN,   focus=sidebar        →  Sidebar CLOSED, focus=market
```

**Test additions for Task 3.5 (app_test.go):**

```go
func TestTabOpensClosedSidebar(t *testing.T) {
    m := New(DefaultTestConfig(), nil)
    m.sidebarOpen = false
    m.focus = focusMarket

    // Press Tab
    updated, _ := m.Update(tea.KeyPressMsg{/* tab */})
    app := updated.(App)
    if !app.sidebarOpen {
        t.Error("expected sidebar to open")
    }
    if app.focus != focusSidebar {
        t.Error("expected focus on sidebar")
    }
}

func TestTabClosesSidebarWhenFocused(t *testing.T) {
    m := New(DefaultTestConfig(), nil)
    m.sidebarOpen = true
    m.focus = focusSidebar

    // Press Tab
    updated, _ := m.Update(tea.KeyPressMsg{/* tab */})
    app := updated.(App)
    if app.sidebarOpen {
        t.Error("expected sidebar to close")
    }
    if app.focus != focusMarket {
        t.Error("expected focus on market")
    }
}

func TestTabFocusesSidebarWhenOpenAndMarketFocused(t *testing.T) {
    m := New(DefaultTestConfig(), nil)
    m.sidebarOpen = true
    m.focus = focusMarket

    // Press Tab
    updated, _ := m.Update(tea.KeyPressMsg{/* tab */})
    app := updated.(App)
    if !app.sidebarOpen {
        t.Error("expected sidebar to stay open")
    }
    if app.focus != focusSidebar {
        t.Error("expected focus on sidebar")
    }
}
```

---

## HIGH FIX 3: Intraday Price Storage for Sparklines

**Problem:** `market_cache` stores only the latest price per symbol (one row). Sparklines need a time series of ~50 intraday price points. Addendum Task 4.6 says "feed intraday price data from cache" but no such data exists.

**Fix:** Use an in-memory ring buffer per symbol. Populated from tick loop price updates. Lost on restart (acceptable for sparklines — they rebuild within minutes of launch).

**Add to Task 4.5 (Portfolio View):**

```go
// internal/ui/market/portfolio/sparkline_buffer.go
package portfolio

import "sync"

const sparklineCapacity = 60 // ~30 min at 30s intervals

// SparklineBuffer holds recent price points for inline sparklines.
type SparklineBuffer struct {
    mu   sync.RWMutex
    data map[string]*ringBuf // symbol → prices
}

type ringBuf struct {
    values []float64
    head   int
    count  int
}

func NewSparklineBuffer() *SparklineBuffer {
    return &SparklineBuffer{data: make(map[string]*ringBuf)}
}

func (sb *SparklineBuffer) Push(symbol string, price float64) {
    sb.mu.Lock()
    defer sb.mu.Unlock()
    buf, ok := sb.data[symbol]
    if !ok {
        buf = &ringBuf{values: make([]float64, sparklineCapacity)}
        sb.data[symbol] = buf
    }
    buf.values[buf.head] = price
    buf.head = (buf.head + 1) % sparklineCapacity
    if buf.count < sparklineCapacity {
        buf.count++
    }
}

// Values returns prices in chronological order.
func (sb *SparklineBuffer) Values(symbol string) []float64 {
    sb.mu.RLock()
    defer sb.mu.RUnlock()
    buf, ok := sb.data[symbol]
    if !ok || buf.count == 0 {
        return nil
    }
    result := make([]float64, buf.count)
    start := (buf.head - buf.count + sparklineCapacity) % sparklineCapacity
    for i := 0; i < buf.count; i++ {
        result[i] = buf.values[(start+i)%sparklineCapacity]
    }
    return result
}
```

**Wire into tick loop (Task 4.7):** After each `QuotesMsg` is received, push prices into the sparkline buffer:

```go
case QuotesMsg:
    for _, q := range msg.Quotes {
        m.sparklines.Push(q.Symbol, q.Price)
    }
    // ... existing cache update logic
```

**Addendum Task 4.6 sparkline step becomes:** Use `ChartRenderer.RenderSparkline(m.sparklines.Values(symbol), width)` for each row.

**Test:**

```go
func TestSparklineBufferOrder(t *testing.T) {
    buf := NewSparklineBuffer()
    for i := 1; i <= 5; i++ {
        buf.Push("AAPL", float64(i))
    }
    vals := buf.Values("AAPL")
    if len(vals) != 5 {
        t.Fatalf("expected 5, got %d", len(vals))
    }
    for i, v := range vals {
        if v != float64(i+1) {
            t.Errorf("index %d: expected %f, got %f", i, float64(i+1), v)
        }
    }
}

func TestSparklineBufferWraparound(t *testing.T) {
    buf := NewSparklineBuffer()
    for i := 0; i < sparklineCapacity+10; i++ {
        buf.Push("BTC-USD", float64(i))
    }
    vals := buf.Values("BTC-USD")
    if len(vals) != sparklineCapacity {
        t.Fatalf("expected %d, got %d", sparklineCapacity, len(vals))
    }
    // First value should be 10 (oldest after wrapping)
    if vals[0] != 10 {
        t.Errorf("expected oldest=10, got %f", vals[0])
    }
}
```

---

## HIGH FIX 4: Candle Type — Prevent Circular Import

**Problem:** `service/market/provider.go` defines `Candle`. `ui/common/chart.go` defines `ChartRenderer`. Adding `RenderCandlestick(candles []Candle, ...)` to the interface would create a `ui/common` → `service/market` import or vice versa.

**Supersedes:** Addendum Task 5.7 — the `ChartRenderer` interface definition.

**Fix:** Define chart-specific types in `ui/common/chart.go`. Service layer maps at the boundary.

```go
// internal/ui/common/chart.go
package common

// Candle is the chart renderer's own candle type.
// Service layer maps market.Candle → common.Candle at the boundary.
type Candle struct {
    Time   int64
    Open   float64
    High   float64
    Low    float64
    Close  float64
    Volume float64
}

type ChartData struct {
    Labels []string
    Values []float64
}

type ChartRenderer interface {
    RenderLineChart(data ChartData, width, height int) string
    RenderCandlestick(candles []Candle, width, height int) string
    RenderSparkline(values []float64, width int) string
}
```

**Mapping helper (in `internal/app/messages.go` or a shared conversion package):**

```go
// convertCandles maps market provider candles to chart renderer candles.
func convertCandles(src []market.Candle) []common.Candle {
    out := make([]common.Candle, len(src))
    for i, c := range src {
        out[i] = common.Candle{
            Time: c.Time, Open: c.Open, High: c.High,
            Low: c.Low, Close: c.Close, Volume: c.Volume,
        }
    }
    return out
}
```

**Dependency direction (import graph):**

```
service/market (defines market.Candle)
    ↓ used by
internal/app (imports both, does the conversion)
    ↓ passes to
ui/common (defines common.Candle, ChartRenderer)
    ↓ used by
ui/market/charts (renders charts)
```

No circular imports. `service/market` and `ui/common` never import each other.

---

## HIGH FIX 5: Migration Runner Version Check

**Problem:** `RunMigrations()` runs ALL `.sql` files every time. Works for `CREATE TABLE IF NOT EXISTS` but breaks on future `ALTER TABLE` migrations.

**Supersedes:** Original plan Task 1.4, Step 5 — the `RunMigrations` function in `store/db.go`.

**Fix:** Check the `schema_migrations` table before each migration.

```go
// internal/store/db.go

func RunMigrations(db *sql.DB) error {
    // Ensure migration tracking table exists
    if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
        version INTEGER PRIMARY KEY,
        applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`); err != nil {
        return fmt.Errorf("create schema_migrations: %w", err)
    }

    // Get current schema version
    var currentVersion int
    err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
    if err != nil {
        return fmt.Errorf("read schema version: %w", err)
    }

    entries, err := migrations.FS.ReadDir(".")
    if err != nil {
        return fmt.Errorf("read migrations: %w", err)
    }

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        name := entry.Name()
        // Extract version number from filename: "001_initial_schema.sql" → 1
        if !strings.HasSuffix(name, ".sql") {
            continue
        }
        versionStr := strings.SplitN(name, "_", 2)[0]
        version, err := strconv.Atoi(versionStr)
        if err != nil {
            slog.Warn("skipping non-versioned migration file", "file", name)
            continue
        }

        if version <= currentVersion {
            continue // already applied
        }

        data, err := migrations.FS.ReadFile(name)
        if err != nil {
            return fmt.Errorf("read %s: %w", name, err)
        }
        if _, err := db.Exec(string(data)); err != nil {
            return fmt.Errorf("exec %s: %w", name, err)
        }

        // Record migration
        if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
            return fmt.Errorf("record migration %d: %w", version, err)
        }
        slog.Info("migration applied", "file", name, "version", version)
    }
    return nil
}
```

**Consequential change to `001_initial_schema.sql`:** Remove the `CREATE TABLE schema_migrations` and `INSERT` from the SQL file since the runner now handles it. The SQL file should only contain application tables.

**Remove from `001_initial_schema.sql`:**
```sql
-- DELETE THESE LINES from the migration file:
-- CREATE TABLE IF NOT EXISTS schema_migrations (
--     version INTEGER PRIMARY KEY,
--     applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
-- );
-- INSERT OR IGNORE INTO schema_migrations (version) VALUES (1);
```

**Update test in `db_test.go`:**

```go
func TestMigrationVersionTracking(t *testing.T) {
    db, err := Open(":memory:")
    if err != nil {
        t.Fatalf("open: %v", err)
    }
    defer db.Close()

    // Check version was recorded
    var version int
    err = db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
    if err != nil {
        t.Fatalf("query version: %v", err)
    }
    if version != 1 {
        t.Errorf("expected version 1, got %d", version)
    }

    // Run again — should be idempotent (no re-execution)
    if err := RunMigrations(db); err != nil {
        t.Fatalf("second run: %v", err)
    }
}
```

---

## HIGH FIX 6: Standup Upsert Preserves `created_at`

**Problem:** `INSERT OR REPLACE` in SQLite deletes the old row and inserts a new one, resetting `created_at`.

**Supersedes:** Original plan Task 2.4, Step 3 — the `UpsertStandup` function.

**Fix:** Use `ON CONFLICT ... DO UPDATE`:

```go
func UpsertStandup(db *sql.DB, date, yesterday, today, blockers string) error {
    _, err := db.Exec(`
        INSERT INTO standup_entries (date, yesterday, today, blockers, created_at, updated_at)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        ON CONFLICT(date) DO UPDATE SET
            yesterday = excluded.yesterday,
            today = excluded.today,
            blockers = excluded.blockers,
            updated_at = CURRENT_TIMESTAMP
    `, date, yesterday, today, blockers)
    return err
}
```

**Add test:**

```go
func TestUpsertPreservesCreatedAt(t *testing.T) {
    db, _ := Open(":memory:")
    defer db.Close()

    // Create entry
    UpsertStandup(db, "2026-03-13", "did A", "will B", "none")
    entry1, _ := GetStandup(db, "2026-03-13")
    createdAt := entry1.CreatedAt

    // Wait a tiny bit to differentiate timestamps
    time.Sleep(10 * time.Millisecond)

    // Update same date
    UpsertStandup(db, "2026-03-13", "did C", "will D", "blocked")
    entry2, _ := GetStandup(db, "2026-03-13")

    if !entry2.CreatedAt.Equal(createdAt) {
        t.Errorf("created_at changed: was %v, now %v", createdAt, entry2.CreatedAt)
    }
    if entry2.Yesterday != "did C" {
        t.Errorf("expected 'did C', got %s", entry2.Yesterday)
    }
    if entry2.UpdatedAt == nil || entry2.UpdatedAt.Before(createdAt) {
        t.Error("updated_at should be set and after created_at")
    }
}
```

---

## MEDIUM FIX 7: CoinGecko Config and API Key

**Problem:** Config struct has no `CoinGeckoConfig`. CoinGecko Demo API requires an API key for most endpoints. No env var override for it.

**Supersedes:** Original plan Task 1.3, Step 3 — the `MarketConfig` struct and `ApplyEnvOverrides`.

**Add to `internal/config/config.go`:**

```go
type MarketConfig struct {
    StockProvider         string           `toml:"stock_provider"`
    CryptoProvider        string           `toml:"crypto_provider"`
    RefreshIntervalMarket string           `toml:"refresh_interval_market"`
    RefreshIntervalCrypto string           `toml:"refresh_interval_crypto"`
    RefreshIntervalClosed string           `toml:"refresh_interval_closed"`
    Finnhub               FinnhubConfig    `toml:"finnhub"`
    CoinGecko             CoinGeckoConfig  `toml:"coingecko"`  // NEW
}

type CoinGeckoConfig struct {
    APIKey string `toml:"api_key"`  // env: HOARD_COINGECKO_KEY (Demo API)
}
```

**Add to `ApplyEnvOverrides()`:**

```go
if v := os.Getenv("HOARD_COINGECKO_KEY"); v != "" {
    c.Market.CoinGecko.APIKey = v
}
```

**Add to `config.example.toml`:**

```toml
[market.coingecko]
api_key = ""  # env: HOARD_COINGECKO_KEY (free at coingecko.com/en/api/pricing)
```

**Add test:**

```go
func TestCoinGeckoEnvOverride(t *testing.T) {
    t.Setenv("HOARD_COINGECKO_KEY", "cg-demo-key")
    cfg := DefaultConfig()
    cfg.ApplyEnvOverrides()
    if cfg.Market.CoinGecko.APIKey != "cg-demo-key" {
        t.Errorf("expected cg-demo-key, got %s", cfg.Market.CoinGecko.APIKey)
    }
}
```

---

## MEDIUM FIX 8: Help Overlay

**Problem:** Footer shows `[?] help`, keybindings define `?`, but no task implements the help screen.

**Fix:** Add Task 3.9 to Phase 3.

### Task 3.9 (NEW): Help Overlay

**Files:**
- Create: `internal/ui/help/model.go`
- Modify: `internal/app/update.go`

- [ ] **Step 1: Implement help overlay model**

```go
// internal/ui/help/model.go
package help

import (
    "strings"

    "charm.land/lipgloss/v2"
)

type Binding struct {
    Key  string
    Desc string
}

var globalBindings = []Binding{
    {"1-4", "Switch market views"},
    {"Tab", "Toggle sidebar"},
    {"/", "Search symbols"},
    {"?", "Toggle this help"},
    {"q", "Quit"},
}

var portfolioBindings = []Binding{
    {"a", "Add position"},
    {"e", "Edit position"},
    {"d", "Delete position"},
    {"f", "Filter by account"},
    {"r", "Refresh prices"},
    {"j/k", "Navigate up/down"},
    {"s", "Sort column"},
}

var sidebarBindings = []Binding{
    {"a", "Add task / edit standup"},
    {"x", "Toggle task done"},
    {"d", "Delete task"},
    {"e", "Edit standup"},
    {"h", "Standup history"},
    {"g", "Generate standup from git"},
}

// Render returns the help overlay content sized to width×height.
func Render(width, height int) string {
    var b strings.Builder
    b.WriteString("  KEYBOARD SHORTCUTS\n\n")
    writeSection(&b, "Global", globalBindings)
    writeSection(&b, "Portfolio", portfolioBindings)
    writeSection(&b, "Sidebar", sidebarBindings)
    b.WriteString("\n  Press ? or Esc to close")
    return b.String()
}

func writeSection(b *strings.Builder, title string, bindings []Binding) {
    b.WriteString("  " + title + "\n")
    for _, bind := range bindings {
        b.WriteString("    " + bind.Key + "  " + bind.Desc + "\n")
    }
    b.WriteString("\n")
}
```

- [ ] **Step 2: Wire into App.Update**

```go
// In App model:
showHelp bool

// In Update(), modeNormal:
case "?":
    m.showHelp = !m.showHelp
    return m, nil

// In View():
if m.showHelp {
    overlay := help.Render(m.width, m.height)
    // Center overlay on screen using lipgloss.Place
    return tea.View{Content: lipgloss.Place(m.width, m.height,
        lipgloss.Center, lipgloss.Center, overlay)}
}
```

- [ ] **Step 3: Escape closes help**

In the global Escape handler (already exists), add: `m.showHelp = false`

- [ ] **Step 4: Commit**

```bash
git add internal/ui/help/ internal/app/
git commit -m "feat(ui): help overlay with keyboard shortcut reference"
```

---

## MEDIUM FIX 9: Cross-Platform Browser Open

**Problem:** `exec.Command("open", url)` is macOS only.

**Fix:** Add a `browser` utility. Used in Task 6.1 (news view `[o]` key).

**Add to Phase 6, before Task 6.1:**

```go
// internal/ui/common/browser.go
package common

import (
    "os/exec"
    "runtime"
)

// OpenURL opens a URL in the user's default browser.
func OpenURL(url string) error {
    switch runtime.GOOS {
    case "darwin":
        return exec.Command("open", url).Start()
    case "linux":
        return exec.Command("xdg-open", url).Start()
    case "windows":
        return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
    default:
        return exec.Command("xdg-open", url).Start()
    }
}
```

**Supersedes:** Task 6.1 Step 3 — replace `exec.Command("open", url)` with `common.OpenURL(url)`.

---

## MEDIUM FIX 10: Remove `doing` Status from Tasks

**Problem:** Schema has `CHECK(status IN ('todo','doing','done'))` but design spec says "Dead simple: title + done/not-done". The `doing` status is never used.

**Supersedes:** Original plan Task 1.4 — the `tasks` table in `001_initial_schema.sql`.

**Fix:** Change the constraint:

```sql
-- BEFORE:
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'todo' CHECK(status IN ('todo','doing','done')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- AFTER:
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'todo' CHECK(status IN ('todo','done')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);
```

No code changes needed — `ToggleTask` already only uses `todo` and `done`.

---

## MEDIUM FIX 11: CoinGecko Symbol Mapping Strategy

**Problem:** Task 4.2 mentions "maintain a small mapping table for top coins; fall back to search API" but doesn't specify the approach.

**Supersedes:** Original plan Task 4.2, Step 3 — adds implementation detail.

**Fix:** Hardcoded Go map for top coins, in-memory cache for discovered mappings, search API fallback.

```go
// internal/service/market/coingecko_symbols.go
package market

// topCoinIDs maps common ticker symbols to CoinGecko IDs.
// Covers the top ~30 coins by market cap.
var topCoinIDs = map[string]string{
    "BTC":   "bitcoin",
    "ETH":   "ethereum",
    "BNB":   "binancecoin",
    "SOL":   "solana",
    "XRP":   "ripple",
    "ADA":   "cardano",
    "DOGE":  "dogecoin",
    "AVAX":  "avalanche-2",
    "DOT":   "polkadot",
    "LINK":  "chainlink",
    "MATIC": "matic-network",
    "SHIB":  "shiba-inu",
    "UNI":   "uniswap",
    "LTC":   "litecoin",
    "ATOM":  "cosmos",
    "NEAR":  "near",
    "APT":   "aptos",
    "ARB":   "arbitrum",
    "OP":    "optimism",
    "FIL":   "filecoin",
    "AAVE":  "aave",
    "MKR":   "maker",
    "GRT":   "the-graph",
    "IMX":   "immutable-x",
    "SAND":  "the-sandbox",
    "MANA":  "decentraland",
    "AXS":   "axie-infinity",
    "CRV":   "curve-dao-token",
    "COMP":  "compound-governance-token",
    "ALGO":  "algorand",
}

// discoveredIDs caches symbol→ID mappings found via the search API.
// Protected by the CoinGecko client's mutex.
var discoveredIDs = make(map[string]string)

// ResolveCoinID maps a ticker symbol (e.g., "BTC") to a CoinGecko ID.
// 1. Check hardcoded map
// 2. Check discovered cache
// 3. Call CoinGecko /search API, cache result
func (c *CoinGeckoClient) ResolveCoinID(ctx context.Context, symbol string) (string, error) {
    // Strip "-USD" suffix if present: "BTC-USD" → "BTC"
    ticker := strings.TrimSuffix(symbol, "-USD")
    ticker = strings.ToUpper(ticker)

    if id, ok := topCoinIDs[ticker]; ok {
        return id, nil
    }
    if id, ok := discoveredIDs[ticker]; ok {
        return id, nil
    }

    // Search API fallback
    results, err := c.client.Search(ctx, ticker)
    if err != nil {
        return "", fmt.Errorf("coingecko search %s: %w", ticker, err)
    }
    if len(results.Coins) > 0 {
        id := results.Coins[0].ID
        discoveredIDs[ticker] = id
        return id, nil
    }
    return "", fmt.Errorf("unknown coin symbol: %s", ticker)
}
```

**Test:**

```go
func TestResolveCoinIDKnown(t *testing.T) {
    client := NewCoinGeckoClient("", nil)
    id, err := client.ResolveCoinID(context.Background(), "BTC-USD")
    if err != nil {
        t.Fatal(err)
    }
    if id != "bitcoin" {
        t.Errorf("expected 'bitcoin', got %s", id)
    }
}

func TestResolveCoinIDStripsSuffix(t *testing.T) {
    client := NewCoinGeckoClient("", nil)
    id, _ := client.ResolveCoinID(context.Background(), "ETH-USD")
    if id != "ethereum" {
        t.Errorf("expected 'ethereum', got %s", id)
    }
}
```

---

## MEDIUM FIX 12: Import Subcommand Architecture

**Problem:** Task 8.2 mentions `hoard import robinhood path/to/file.csv` but Task 1.8 uses stdlib `flag` which doesn't support subcommands.

**Fix:** Simple `os.Args` routing before flag parsing. Only one subcommand for MVP — no need for a framework.

**Supersedes:** Original plan Task 1.8 — update `main.go` structure.

```go
// cmd/hoard/main.go

func main() {
    // Check for subcommands BEFORE flag parsing
    if len(os.Args) > 1 && os.Args[1] == "import" {
        runImport(os.Args[2:])
        return
    }

    // Normal TUI mode
    flag.Parse()
    // ... rest of existing main()
}

// runImport handles: hoard import <format> <file>
func runImport(args []string) {
    if len(args) < 2 {
        fmt.Fprintf(os.Stderr, "Usage: hoard import <format> <file>\nFormats: robinhood\n")
        os.Exit(1)
    }
    format, filepath := args[0], args[1]

    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "config error: %v\n", err)
        os.Exit(1)
    }
    if err := config.EnsureDirs(); err != nil {
        fmt.Fprintf(os.Stderr, "dir error: %v\n", err)
        os.Exit(1)
    }
    db, err := store.Open(config.DBFilePath())
    if err != nil {
        fmt.Fprintf(os.Stderr, "db error: %v\n", err)
        os.Exit(1)
    }
    defer db.Close()

    switch format {
    case "robinhood":
        count, err := importer.ImportRobinhoodCSV(db, filepath, cfg.BaseCurrency)
        if err != nil {
            fmt.Fprintf(os.Stderr, "import error: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("Imported %d transactions from Robinhood CSV.\n", count)
    default:
        fmt.Fprintf(os.Stderr, "Unknown format: %s\nSupported: robinhood\n", format)
        os.Exit(1)
    }
}
```

**Also remove the in-app `[i]mport` key from Task 8.2 Step 5.** Import is CLI-only, not TUI. Reason: file picker in a TUI is complex and the user already has the file path. CLI is simpler and more reliable.

---

## MEDIUM FIX 13: Charts View Symbol Selection

**Problem:** Charts view shows a chart for a "selected symbol" but no UI mechanism is described for picking which symbol to chart.

**Supersedes:** Original plan Task 5.3 — adds symbol selection behavior.

**Fix:** Two mechanisms:

**A) Auto-select from portfolio/watchlist (primary):**
When switching to Charts view (pressing `3`), auto-chart the symbol that was highlighted in the previous view (Portfolio or Watchlist). Store the "last selected symbol" at the App level.

```go
// In App model:
selectedSymbol string  // set by portfolio/watchlist row selection

// In update.go, when switching to charts:
case "3":
    m.activeView = viewCharts
    if m.selectedSymbol != "" {
        // Trigger chart data fetch for the selected symbol
        return m, m.fetchCandles(m.selectedSymbol, m.charts.Timeframe())
    }
```

**B) Symbol picker in Charts view (secondary):**
Press `s` in Charts view to open a text input for entering a symbol directly (switches to `modeTextInput`). Uses the same search logic as the search overlay.

**Add to Task 5.3 tests:**

```go
func TestChartsAutoSelectsSymbol(t *testing.T) {
    m := newTestApp()
    m.selectedSymbol = "AAPL"
    // Switch to charts
    updated, cmd := m.Update(keyMsg("3"))
    app := updated.(App)
    if app.charts.Symbol() != "AAPL" {
        t.Errorf("expected AAPL, got %s", app.charts.Symbol())
    }
    if cmd == nil {
        t.Error("expected candle fetch command")
    }
}
```

---

## Updated Phase Task Counts

| Phase | Name | Tasks | Change from v2 |
|-------|------|-------|----------------|
| 1 | Foundation | 10 | unchanged (code fixes within existing tasks) |
| 1.5 | CI/CD Pipeline | 3 | unchanged |
| 2 | Data Layer | 5 | unchanged (code fixes within existing tasks) |
| 3 | App Shell & Layout | **9 (+1)** | Added: Task 3.9 (help overlay) |
| 4 | Portfolio & Market Data | 9 | unchanged (sparkline buffer added to existing task) |
| 5 | Watchlist & Charts | 7 | unchanged (symbol selection added to existing task) |
| 6 | News & Sidebar | 7 | unchanged (browser.go added to existing code) |
| 6.5 | Distribution & Packaging | 4 | unchanged |
| 7 | AI Layer | 6 | unchanged |
| 8 | Polish & Release | **5 (-1)** | Task 8.4 SUPERSEDED by Phase 6.5 |

**Total: 65 tasks** (net unchanged: +1 new, -1 superseded)

---

## Summary of All v3 Changes

### HIGH Fixes (6)

| # | Issue | Fix | Affects |
|---|-------|-----|---------|
| 1 | GoReleaser defined twice | Mark Task 8.4 as SUPERSEDED | Phase 8 |
| 2 | Tab only opens sidebar | Three-state machine with tests | Task 1.5 / Task 3.5 |
| 3 | No intraday data for sparklines | In-memory ring buffer per symbol | Tasks 4.5, 4.6, 4.7 |
| 4 | Candle type circular import | Define `common.Candle`, map at boundary | Task 5.2, 5.7 |
| 5 | Migration runner re-runs all | Version check + skip applied | Task 1.4 |
| 6 | Standup upsert resets `created_at` | `ON CONFLICT DO UPDATE` | Task 2.4 |

### MEDIUM Fixes (7)

| # | Issue | Fix | Affects |
|---|-------|-----|---------|
| 7 | Missing CoinGecko config | Add `CoinGeckoConfig` + env var | Task 1.3 |
| 8 | Help overlay missing | New Task 3.9 | Phase 3 |
| 9 | macOS-only browser open | `common.OpenURL()` helper | Task 6.1 |
| 10 | Unused `doing` task status | Remove from CHECK constraint | Task 1.4 |
| 11 | CoinGecko symbol mapping vague | Hardcoded map + search fallback | Task 4.2 |
| 12 | Import subcommand undefined | `os.Args` routing, CLI-only import | Tasks 1.8, 8.2 |
| 13 | Charts symbol selection undefined | Auto-select from prev view + `s` picker | Task 5.3 |
