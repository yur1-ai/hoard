# Hoard Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a terminal-based personal finance dashboard (stocks, ETFs, crypto) with a collapsible daily cockpit sidebar (calendar, tasks, standup) in Go + Bubble Tea v2.

**Architecture:** Single-binary TUI app using the Elm Architecture (Bubble Tea v2). SQLite for persistence, provider interfaces for all external services (market data, calendar, AI) enabling graceful degradation. Layout: full-screen market dashboard with four sub-views, collapsible right sidebar with three panels.

**Tech Stack:** Go 1.23+, Bubble Tea v2 (`charm.land/bubbletea/v2`), Lip Gloss v2, Bubbles v2, modernc.org/sqlite (pure Go), TOML config, Finnhub (stocks), CoinGecko (crypto), Frankfurter (currency), Groq/Gemini/Ollama (AI).

**Spec:** `docs/design-spec.md` and `docs/architecture.md`

---

## Phase Overview

| Phase | Name | Tasks | What You'll Have After |
|-------|------|-------|----------------------|
| 1 | Foundation | 6 | Go module, config, SQLite, empty TUI shell that starts/quits |
| 2 | Data Layer | 5 | Full CRUD store for portfolio, watchlists, tasks, standups, cache |
| 3 | App Shell & Layout | 6 | Multi-panel layout, sidebar toggle, view switching, input modes |
| 4 | Portfolio & Market Data | 7 | Live portfolio view with real stock/crypto prices from APIs |
| 5 | Watchlist & Charts | 5 | Watchlist management + price charts with timeframes |
| 6 | News & Sidebar | 6 | News feed + calendar, tasks, standup in sidebar |
| 7 | AI Layer | 5 | Optional AI features with provider chain + graceful fallbacks |
| 8 | Polish & Release | 5 | CSV import, first-run UX, error display, Makefile, GoReleaser |

**Total: 45 tasks**

---

## File Map

Files created across all phases. Each file has one clear responsibility.

```
hoard/
├── cmd/hoard/
│   └── main.go                          # Phase 1: entry point
├── internal/
│   ├── config/
│   │   ├── config.go                    # Phase 1: TOML + env var loading
│   │   ├── config_test.go               # Phase 1: config tests
│   │   └── paths.go                     # Phase 1: XDG path resolution
│   ├── store/
│   │   ├── db.go                        # Phase 1: SQLite init + migrations
│   │   ├── db_test.go                   # Phase 1: migration tests
│   │   ├── portfolio.go                 # Phase 2: accounts + holdings + transactions
│   │   ├── portfolio_test.go            # Phase 2
│   │   ├── watchlist.go                 # Phase 2: watchlist CRUD
│   │   ├── watchlist_test.go            # Phase 2
│   │   ├── tasks.go                     # Phase 2: simple tasks CRUD
│   │   ├── tasks_test.go               # Phase 2
│   │   ├── standup.go                   # Phase 2: standup entries CRUD
│   │   ├── standup_test.go             # Phase 2
│   │   ├── cache.go                     # Phase 2: market price cache
│   │   └── cache_test.go               # Phase 2
│   ├── app/
│   │   ├── model.go                     # Phase 3: root model, Init, layout state
│   │   ├── update.go                    # Phase 3: message routing, input mode
│   │   ├── view.go                      # Phase 3: layout composition
│   │   ├── messages.go                  # Phase 3: all message types
│   │   └── app_test.go                  # Phase 3: model/update tests
│   ├── ui/
│   │   ├── common/
│   │   │   ├── styles.go               # Phase 3: Lip Gloss theme
│   │   │   ├── keys.go                 # Phase 3: keybinding definitions
│   │   │   └── chart.go                # Phase 5: ChartRenderer interface
│   │   ├── header/
│   │   │   └── model.go                # Phase 3: header bar
│   │   ├── footer/
│   │   │   └── model.go                # Phase 3: footer bar + help
│   │   ├── market/
│   │   │   ├── portfolio/
│   │   │   │   ├── model.go            # Phase 4: holdings table + P&L
│   │   │   │   ├── view.go             # Phase 4: table rendering
│   │   │   │   └── model_test.go       # Phase 4
│   │   │   ├── watchlist/
│   │   │   │   ├── model.go            # Phase 5: named lists
│   │   │   │   ├── view.go             # Phase 5
│   │   │   │   └── model_test.go       # Phase 5
│   │   │   ├── charts/
│   │   │   │   ├── model.go            # Phase 5: chart data + timeframe
│   │   │   │   ├── view.go             # Phase 5: chart rendering
│   │   │   │   └── model_test.go       # Phase 5
│   │   │   └── news/
│   │   │       ├── model.go            # Phase 6: filtered articles
│   │   │       ├── view.go             # Phase 6
│   │   │       └── model_test.go       # Phase 6
│   │   └── sidebar/
│   │       ├── model.go                # Phase 3: sidebar container
│   │       ├── calendar/
│   │       │   └── model.go            # Phase 6: calendar agenda
│   │       ├── tasks/
│   │       │   ├── model.go            # Phase 6: task list UI
│   │       │   └── model_test.go       # Phase 6
│   │       └── standup/
│   │           ├── model.go            # Phase 6: standup editor
│   │           └── model_test.go       # Phase 6
│   ├── service/
│   │   ├── market/
│   │   │   ├── provider.go             # Phase 4: MarketProvider interface
│   │   │   ├── finnhub.go              # Phase 4: Finnhub REST client
│   │   │   ├── finnhub_test.go         # Phase 4
│   │   │   ├── coingecko.go            # Phase 4: CoinGecko client
│   │   │   ├── coingecko_test.go       # Phase 4
│   │   │   └── cache.go                # Phase 4: caching layer
│   │   ├── calendar/
│   │   │   ├── provider.go             # Phase 6: CalendarProvider interface
│   │   │   ├── ics.go                  # Phase 6: .ics file parser
│   │   │   ├── ics_test.go             # Phase 6
│   │   │   └── gws.go                  # Phase 6: Google Workspace CLI
│   │   ├── ai/
│   │   │   ├── provider.go             # Phase 7: AIProvider interface
│   │   │   ├── chain.go                # Phase 7: fallback chain
│   │   │   ├── chain_test.go           # Phase 7
│   │   │   ├── ollama.go               # Phase 7
│   │   │   ├── groq.go                 # Phase 7
│   │   │   ├── gemini.go               # Phase 7
│   │   │   └── noop.go                 # Phase 7: no-AI fallback
│   │   └── currency/
│   │       ├── frankfurter.go           # Phase 4: currency rates
│   │       └── frankfurter_test.go      # Phase 4
│   └── importer/
│       ├── importer.go                  # Phase 8: Importer interface
│       ├── robinhood.go                 # Phase 8: Robinhood CSV parser
│       └── robinhood_test.go            # Phase 8
├── migrations/
│   ├── 001_initial_schema.sql           # Phase 1
│   └── embed.go                         # Phase 1
├── config.example.toml                  # Phase 1
├── Makefile                             # Phase 8
├── .goreleaser.yaml                     # Phase 8
├── go.mod                               # Phase 1
└── go.sum                               # Phase 1
```

---

## Chunk 1: Phase 1 — Foundation

**Goal:** Bootable app skeleton — Go module, config loading, SQLite with migrations, empty Bubble Tea shell that starts and quits with `q`.

### Task 1.1: Initialize Go Module and Dependencies

**Files:**
- Create: `go.mod`
- Create: `cmd/hoard/main.go` (minimal)

- [ ] **Step 1: Create project directory and init Go module**

```bash
cd ~/ai-projects/hoard
go mod init github.com/yourusername/hoard
```

- [ ] **Step 2: Install core dependencies**

```bash
# Bubble Tea v2 ecosystem
go get charm.land/bubbletea/v2
go get charm.land/lipgloss/v2
go get charm.land/bubbles/v2
go get charm.land/glamour@latest

# SQLite (pure Go, no CGO)
go get modernc.org/sqlite

# Config
go get github.com/pelletier/go-toml/v2

# Market data
go get github.com/Finnhub-Stock-API/finnhub-go/v2
go get github.com/JulianToledano/goingecko/v3
```

- [ ] **Step 3: Create minimal main.go**

```go
// cmd/hoard/main.go
package main

import (
    "fmt"
    "os"
)

var version = "dev"

func main() {
    fmt.Fprintf(os.Stderr, "hoard %s\n", version)
    os.Exit(0)
}
```

- [ ] **Step 4: Verify it builds and runs**

```bash
go build -o bin/hoard ./cmd/hoard
./bin/hoard
# Expected: "hoard dev"
```

- [ ] **Step 5: Commit**

```bash
git init
echo "bin/" > .gitignore
echo "*.db" >> .gitignore
git add go.mod go.sum cmd/ .gitignore
git commit -m "chore: init Go module with core dependencies"
```

---

### Task 1.2: XDG Path Resolution

**Files:**
- Create: `internal/config/paths.go`
- Create: `internal/config/paths_test.go` (implicitly tested via config tests)

- [ ] **Step 1: Write paths.go with XDG-compliant path functions**

```go
// internal/config/paths.go
package config

import (
    "os"
    "path/filepath"
)

// ConfigDir returns ~/.config/hoard (or XDG_CONFIG_HOME/hoard).
func ConfigDir() string {
    if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
        return filepath.Join(xdg, "hoard")
    }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "hoard")
}

// DataDir returns ~/.local/share/hoard (or XDG_DATA_HOME/hoard).
func DataDir() string {
    if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
        return filepath.Join(xdg, "hoard")
    }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".local", "share", "hoard")
}

// ConfigFilePath returns the full path to config.toml.
func ConfigFilePath() string {
    return filepath.Join(ConfigDir(), "config.toml")
}

// DBFilePath returns the full path to hoard.db.
func DBFilePath() string {
    return filepath.Join(DataDir(), "hoard.db")
}

// LogFilePath returns the full path to hoard.log.
func LogFilePath() string {
    return filepath.Join(DataDir(), "hoard.log")
}

// EnsureDirs creates config and data directories if they don't exist.
func EnsureDirs() error {
    if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
        return err
    }
    return os.MkdirAll(DataDir(), 0755)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/config/
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add internal/config/paths.go
git commit -m "feat(config): add XDG-compliant path resolution"
```

---

### Task 1.3: Config Loading (TOML + Env Vars)

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `config.example.toml`

- [ ] **Step 1: Write the failing test**

```go
// internal/config/config_test.go
package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadDefaults(t *testing.T) {
    cfg := DefaultConfig()
    if cfg.BaseCurrency != "USD" {
        t.Errorf("expected USD, got %s", cfg.BaseCurrency)
    }
    if cfg.Theme != "dark" {
        t.Errorf("expected dark, got %s", cfg.Theme)
    }
    if cfg.SidebarDefault != "open" {
        t.Errorf("expected open, got %s", cfg.SidebarDefault)
    }
}

func TestLoadFromTOML(t *testing.T) {
    dir := t.TempDir()
    toml := `
base_currency = "EUR"
theme = "light"
sidebar_default = "closed"

[market]
refresh_interval_market = "15s"

[market.finnhub]
api_key = "test-key-123"

[ai]
provider_priority = ["groq", "ollama"]
timeout_ms = 3000
`
    path := filepath.Join(dir, "config.toml")
    os.WriteFile(path, []byte(toml), 0644)

    cfg, err := LoadFromFile(path)
    if err != nil {
        t.Fatalf("load error: %v", err)
    }
    if cfg.BaseCurrency != "EUR" {
        t.Errorf("expected EUR, got %s", cfg.BaseCurrency)
    }
    if cfg.Theme != "light" {
        t.Errorf("expected light, got %s", cfg.Theme)
    }
    if cfg.Market.Finnhub.APIKey != "test-key-123" {
        t.Errorf("expected test-key-123, got %s", cfg.Market.Finnhub.APIKey)
    }
    if cfg.AI.TimeoutMs != 3000 {
        t.Errorf("expected 3000, got %d", cfg.AI.TimeoutMs)
    }
}

func TestEnvVarOverride(t *testing.T) {
    t.Setenv("HOARD_FINNHUB_KEY", "env-key-456")
    cfg := DefaultConfig()
    cfg.ApplyEnvOverrides()
    if cfg.Market.Finnhub.APIKey != "env-key-456" {
        t.Errorf("expected env-key-456, got %s", cfg.Market.Finnhub.APIKey)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/config/ -v
# Expected: FAIL — Config, LoadFromFile, ApplyEnvOverrides not defined
```

- [ ] **Step 3: Write config.go implementation**

```go
// internal/config/config.go
package config

import (
    "os"

    toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
    BaseCurrency   string `toml:"base_currency"`
    Theme          string `toml:"theme"`
    SidebarDefault string `toml:"sidebar_default"`

    Market   MarketConfig   `toml:"market"`
    Calendar CalendarConfig `toml:"calendar"`
    AI       AIConfig       `toml:"ai"`
}

type MarketConfig struct {
    StockProvider         string        `toml:"stock_provider"`
    CryptoProvider        string        `toml:"crypto_provider"`
    RefreshIntervalMarket string        `toml:"refresh_interval_market"`
    RefreshIntervalCrypto string        `toml:"refresh_interval_crypto"`
    RefreshIntervalClosed string        `toml:"refresh_interval_closed"`
    Finnhub               FinnhubConfig `toml:"finnhub"`
}

type FinnhubConfig struct {
    APIKey string `toml:"api_key"`
}

type CalendarConfig struct {
    Source  string `toml:"source"`
    ICSPath string `toml:"ics_path"`
}

type AIConfig struct {
    ProviderPriority []string     `toml:"provider_priority"`
    TimeoutMs        int          `toml:"timeout_ms"`
    CacheTTLMinutes  int          `toml:"cache_ttl_minutes"`
    Ollama           OllamaConfig `toml:"ollama"`
    Groq             GroqConfig   `toml:"groq"`
    Gemini           GeminiConfig `toml:"gemini"`
}

type OllamaConfig struct {
    Endpoint string `toml:"endpoint"`
    Model    string `toml:"model"`
}

type GroqConfig struct {
    APIKey string `toml:"api_key"`
}

type GeminiConfig struct {
    APIKey string `toml:"api_key"`
}

func DefaultConfig() Config {
    return Config{
        BaseCurrency:   "USD",
        Theme:          "dark",
        SidebarDefault: "open",
        Market: MarketConfig{
            StockProvider:         "finnhub",
            CryptoProvider:        "coingecko",
            RefreshIntervalMarket: "30s",
            RefreshIntervalCrypto: "120s",
            RefreshIntervalClosed: "5m",
        },
        Calendar: CalendarConfig{
            Source: "auto",
        },
        AI: AIConfig{
            ProviderPriority: []string{"ollama", "groq", "gemini"},
            TimeoutMs:        5000,
            CacheTTLMinutes:  60,
            Ollama: OllamaConfig{
                Endpoint: "http://localhost:11434",
                Model:    "llama3.3",
            },
        },
    }
}

func LoadFromFile(path string) (Config, error) {
    cfg := DefaultConfig()
    data, err := os.ReadFile(path)
    if err != nil {
        return cfg, err
    }
    if err := toml.Unmarshal(data, &cfg); err != nil {
        return cfg, err
    }
    cfg.ApplyEnvOverrides()
    return cfg, nil
}

// Load tries the default config path; returns defaults if file missing.
func Load() (Config, error) {
    path := ConfigFilePath()
    cfg, err := LoadFromFile(path)
    if os.IsNotExist(err) {
        cfg = DefaultConfig()
        cfg.ApplyEnvOverrides()
        return cfg, nil
    }
    return cfg, err
}

func (c *Config) ApplyEnvOverrides() {
    if v := os.Getenv("HOARD_FINNHUB_KEY"); v != "" {
        c.Market.Finnhub.APIKey = v
    }
    if v := os.Getenv("HOARD_GROQ_KEY"); v != "" {
        c.AI.Groq.APIKey = v
    }
    if v := os.Getenv("HOARD_GEMINI_KEY"); v != "" {
        c.AI.Gemini.APIKey = v
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/config/ -v
# Expected: PASS (all 3 tests)
```

- [ ] **Step 5: Create config.example.toml**

Create `config.example.toml` at project root — copy the full TOML from `docs/architecture.md` Section 12, with empty API keys and comments.

- [ ] **Step 6: Commit**

```bash
git add internal/config/ config.example.toml
git commit -m "feat(config): TOML config loading with env var overrides and XDG paths"
```

---

### Task 1.4: SQLite Initialization and Migrations

**Files:**
- Create: `internal/store/db.go`
- Create: `internal/store/db_test.go`
- Create: `migrations/001_initial_schema.sql`
- Create: `migrations/embed.go`

- [ ] **Step 1: Write migration SQL**

```sql
-- migrations/001_initial_schema.sql
CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('brokerage','retirement','crypto_wallet')),
    currency TEXT NOT NULL DEFAULT 'USD',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS holdings (
    id INTEGER PRIMARY KEY,
    account_id INTEGER REFERENCES accounts(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    quantity REAL NOT NULL,
    avg_cost_basis REAL NOT NULL,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY,
    account_id INTEGER REFERENCES accounts(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    type TEXT NOT NULL CHECK(type IN ('buy','sell','dividend','transfer')),
    quantity REAL NOT NULL,
    price REAL NOT NULL,
    fee REAL DEFAULT 0,
    date DATETIME NOT NULL,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS watchlists (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS watchlist_items (
    watchlist_id INTEGER REFERENCES watchlists(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (watchlist_id, symbol)
);

CREATE TABLE IF NOT EXISTS market_cache (
    symbol TEXT PRIMARY KEY,
    market TEXT NOT NULL,
    price REAL,
    change REAL,
    change_pct REAL,
    volume REAL,
    high_24h REAL,
    low_24h REAL,
    last_updated DATETIME
);

CREATE TABLE IF NOT EXISTS currency_rates (
    from_currency TEXT NOT NULL,
    to_currency TEXT NOT NULL,
    rate REAL NOT NULL,
    fetched_at DATETIME NOT NULL,
    PRIMARY KEY (from_currency, to_currency)
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'todo' CHECK(status IN ('todo','doing','done')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

CREATE TABLE IF NOT EXISTS standup_entries (
    id INTEGER PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    yesterday TEXT,
    today TEXT,
    blockers TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS calendar_cache (
    event_id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    calendar_name TEXT,
    last_synced DATETIME
);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT OR IGNORE INTO schema_migrations (version) VALUES (1);
```

- [ ] **Step 2: Write embed.go**

```go
// migrations/embed.go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

- [ ] **Step 3: Write failing test for DB init**

```go
// internal/store/db_test.go
package store

import (
    "testing"
)

func TestOpenInMemory(t *testing.T) {
    db, err := Open(":memory:")
    if err != nil {
        t.Fatalf("failed to open: %v", err)
    }
    defer db.Close()

    // Verify tables exist by querying each one
    tables := []string{
        "accounts", "holdings", "transactions",
        "watchlists", "watchlist_items", "market_cache",
        "currency_rates", "tasks", "standup_entries",
        "calendar_cache", "schema_migrations",
    }
    for _, table := range tables {
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
        if err != nil {
            t.Errorf("table %s not found: %v", table, err)
        }
    }
}

func TestMigrationIdempotent(t *testing.T) {
    db, err := Open(":memory:")
    if err != nil {
        t.Fatalf("first open: %v", err)
    }
    // Run migrations again — should not error
    if err := RunMigrations(db); err != nil {
        t.Fatalf("second migration run failed: %v", err)
    }
    db.Close()
}
```

- [ ] **Step 4: Run tests to verify they fail**

```bash
go test ./internal/store/ -v
# Expected: FAIL — Open, RunMigrations not defined
```

- [ ] **Step 5: Write db.go implementation**

```go
// internal/store/db.go
package store

import (
    "database/sql"
    "fmt"
    "log"

    "github.com/yourusername/hoard/migrations"
    _ "modernc.org/sqlite"
)

// Open opens a SQLite database and runs migrations.
func Open(dsn string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }

    // Enable WAL mode and foreign keys
    pragmas := []string{
        "PRAGMA journal_mode=WAL",
        "PRAGMA foreign_keys=ON",
        "PRAGMA busy_timeout=5000",
    }
    for _, p := range pragmas {
        if _, err := db.Exec(p); err != nil {
            db.Close()
            return nil, fmt.Errorf("pragma %s: %w", p, err)
        }
    }

    if err := RunMigrations(db); err != nil {
        db.Close()
        return nil, fmt.Errorf("migrations: %w", err)
    }

    return db, nil
}

// RunMigrations applies all embedded SQL migration files.
func RunMigrations(db *sql.DB) error {
    entries, err := migrations.FS.ReadDir(".")
    if err != nil {
        return fmt.Errorf("read migrations: %w", err)
    }
    for _, entry := range entries {
        if entry.IsDir() || entry.Name() == "embed.go" {
            continue
        }
        data, err := migrations.FS.ReadFile(entry.Name())
        if err != nil {
            return fmt.Errorf("read %s: %w", entry.Name(), err)
        }
        if _, err := db.Exec(string(data)); err != nil {
            return fmt.Errorf("exec %s: %w", entry.Name(), err)
        }
        log.Printf("migration applied: %s", entry.Name())
    }
    return nil
}

// WALCheckpoint runs a WAL checkpoint to prevent unbounded WAL growth.
func WALCheckpoint(db *sql.DB) error {
    _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
    return err
}
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
go test ./internal/store/ -v
# Expected: PASS (both tests)
```

- [ ] **Step 7: Commit**

```bash
git add internal/store/db.go internal/store/db_test.go migrations/
git commit -m "feat(store): SQLite init with embedded migrations and WAL mode"
```

---

### Task 1.5: Minimal Bubble Tea App Shell

**Files:**
- Create: `internal/app/model.go`
- Create: `internal/app/messages.go`
- Modify: `cmd/hoard/main.go`

- [ ] **Step 1: Write messages.go with core message types**

```go
// internal/app/messages.go
package app

import "time"

// TickMsg fires on each refresh interval.
type TickMsg time.Time

// ErrMsg carries a non-fatal error to display in the status bar.
type ErrMsg struct {
    Err     error
    Context string
}

func (e ErrMsg) Error() string { return e.Err.Error() }
```

- [ ] **Step 2: Write model.go with minimal App model**

```go
// internal/app/model.go
package app

import (
    "database/sql"

    tea "charm.land/bubbletea/v2"
    "github.com/yourusername/hoard/internal/config"
)

type inputMode int

const (
    modeNormal inputMode = iota
    modeTextInput
    modeSearch
)

type activeView int

const (
    viewPortfolio activeView = iota
    viewWatchlist
    viewCharts
    viewNews
)

type focusArea int

const (
    focusMarket focusArea = iota
    focusSidebar
)

// App is the root Bubble Tea model.
type App struct {
    cfg    config.Config
    db     *sql.DB
    width  int
    height int

    mode        inputMode
    activeView  activeView
    focus       focusArea
    sidebarOpen bool

    lastErr string
}

func New(cfg config.Config, db *sql.DB) App {
    sidebarOpen := cfg.SidebarDefault == "open"
    return App{
        cfg:         cfg,
        db:          db,
        mode:        modeNormal,
        activeView:  viewPortfolio,
        focus:       focusMarket,
        sidebarOpen: sidebarOpen,
    }
}

func (m App) Init() tea.Cmd {
    return nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil

    case tea.KeyPressMsg:
        if msg.Code == "escape" {
            m.mode = modeNormal
            return m, nil
        }
        if m.mode == modeNormal {
            switch msg.String() {
            case "q", "ctrl+c":
                return m, tea.Quit
            case "1":
                m.activeView = viewPortfolio
            case "2":
                m.activeView = viewWatchlist
            case "3":
                m.activeView = viewCharts
            case "4":
                m.activeView = viewNews
            case "tab":
                if m.focus == focusMarket {
                    m.focus = focusSidebar
                } else {
                    m.focus = focusMarket
                }
                m.sidebarOpen = true
            }
        }
        return m, nil
    }
    return m, nil
}

func (m App) View() tea.View {
    viewNames := [4]string{"Portfolio", "Watchlist", "Charts", "News"}
    content := "HOARD " + viewNames[m.activeView]
    content += "\n\nPress [1-4] to switch views, [Tab] to toggle sidebar, [q] to quit"

    if m.sidebarOpen {
        content += "\n\n[Sidebar: open]"
    }

    return tea.View{Content: content}
}
```

- [ ] **Step 3: Update main.go to launch the TUI**

```go
// cmd/hoard/main.go
package main

import (
    "fmt"
    "os"

    tea "charm.land/bubbletea/v2"
    "github.com/yourusername/hoard/internal/app"
    "github.com/yourusername/hoard/internal/config"
    "github.com/yourusername/hoard/internal/store"
)

var version = "dev"

func main() {
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

    model := app.New(cfg, db)
    p := tea.NewProgram(model, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

- [ ] **Step 4: Build and manually test**

```bash
go build -o bin/hoard ./cmd/hoard && ./bin/hoard
# Expected: alt-screen TUI showing "HOARD Portfolio"
# Press 1-4 to switch view labels, Tab shows sidebar note, q quits
```

- [ ] **Step 5: Commit**

```bash
git add internal/app/ cmd/hoard/main.go
git commit -m "feat: minimal Bubble Tea v2 app shell with view switching and sidebar toggle"
```

---

### Task 1.6: Makefile and .gitignore

**Files:**
- Create: `Makefile`
- Modify: `.gitignore`

- [ ] **Step 1: Write Makefile**

```makefile
# Makefile
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint clean run

build:
	go build $(LDFLAGS) -o bin/hoard ./cmd/hoard

run: build
	./bin/hoard

test:
	go test ./... -race -count=1

lint:
	golangci-lint run

clean:
	rm -rf bin/
```

- [ ] **Step 2: Update .gitignore**

```
bin/
*.db
*.db-wal
*.db-shm
*.log
.env
```

- [ ] **Step 3: Verify `make build` and `make test` work**

```bash
make build && make test
# Expected: build succeeds, tests pass
```

- [ ] **Step 4: Commit**

```bash
git add Makefile .gitignore
git commit -m "chore: add Makefile and gitignore"
```

---

**Phase 1 complete.** You now have: Go module, TOML config with env overrides, SQLite with all tables, a running Bubble Tea v2 shell with view switching and sidebar toggle, and a Makefile.

---

## Chunk 2: Phase 2 — Data Layer (Store)

**Goal:** Full CRUD store layer for all entities. Pure business logic, no UI, fully tested.

### Task 2.1: Portfolio Store (Accounts + Holdings + Transactions)

**Files:**
- Create: `internal/store/portfolio.go`
- Create: `internal/store/portfolio_test.go`

- [ ] **Step 1: Write failing tests for account and holding CRUD**

Test: create account, create holding, list holdings, get portfolio value, update holding quantity, delete holding. Also test creating a transaction and verifying it recalculates avg cost basis.

Key test cases:
```go
func TestCreateAccountAndHolding(t *testing.T)     // create account, add holding, verify
func TestListHoldings(t *testing.T)                 // multiple holdings, list all, list by account
func TestPortfolioValue(t *testing.T)               // sum of (qty * price) across holdings
func TestAddTransaction(t *testing.T)               // buy transaction updates holding quantity + avg cost
func TestSellTransaction(t *testing.T)              // sell reduces quantity
func TestDeleteHolding(t *testing.T)                // delete removes from DB
```

Each test should use `store.Open(":memory:")` for isolation.

- [ ] **Step 2: Run tests — verify they fail**

```bash
go test ./internal/store/ -run TestCreate -v
# Expected: FAIL — functions not defined
```

- [ ] **Step 3: Implement portfolio.go**

Key types and functions:
```go
type Account struct { ID int64; Name string; Type string; Currency string }
type Holding struct { ID int64; AccountID int64; Symbol string; Market string; Quantity float64; AvgCostBasis float64; Notes string }
type Transaction struct { ID int64; AccountID int64; Symbol string; Market string; Type string; Quantity float64; Price float64; Fee float64; Date time.Time; Notes string }

func CreateAccount(db *sql.DB, name, typ, currency string) (int64, error)
func ListAccounts(db *sql.DB) ([]Account, error)
func CreateHolding(db *sql.DB, h Holding) (int64, error)
func ListHoldings(db *sql.DB, accountID int64) ([]Holding, error)  // 0 = all accounts
func UpdateHolding(db *sql.DB, h Holding) error
func DeleteHolding(db *sql.DB, id int64) error
func AddTransaction(db *sql.DB, tx Transaction) (int64, error)
func ListTransactions(db *sql.DB, symbol string) ([]Transaction, error)
func AllEquitySymbols(db *sql.DB) ([]string, error)
func AllCryptoSymbols(db *sql.DB) ([]string, error)
```

`AddTransaction` should update the corresponding holding's quantity and avg_cost_basis:
- Buy: `new_avg = ((old_qty * old_avg) + (new_qty * price)) / (old_qty + new_qty)`
- Sell: quantity decreases, avg_cost_basis unchanged

- [ ] **Step 4: Run tests — verify they pass**

```bash
go test ./internal/store/ -run TestCreate -v && go test ./internal/store/ -run TestList -v && go test ./internal/store/ -run TestAdd -v && go test ./internal/store/ -run TestSell -v && go test ./internal/store/ -run TestDelete -v
# Expected: all PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/store/portfolio.go internal/store/portfolio_test.go
git commit -m "feat(store): portfolio CRUD — accounts, holdings, transactions with avg cost calculation"
```

---

### Task 2.2: Watchlist Store

**Files:**
- Create: `internal/store/watchlist.go`
- Create: `internal/store/watchlist_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestCreateWatchlist(t *testing.T)            // create, verify name
func TestAddWatchlistItem(t *testing.T)           // add symbol, verify
func TestListWatchlists(t *testing.T)             // multiple lists
func TestListWatchlistItems(t *testing.T)         // items in a specific list
func TestRemoveWatchlistItem(t *testing.T)        // remove symbol
func TestDeleteWatchlist(t *testing.T)            // cascade deletes items
func TestAllWatchedSymbols(t *testing.T)          // unique symbols across all lists
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement watchlist.go**

```go
type Watchlist struct { ID int64; Name string }
type WatchlistItem struct { WatchlistID int64; Symbol string; Market string; AddedAt time.Time }

func CreateWatchlist(db *sql.DB, name string) (int64, error)
func ListWatchlists(db *sql.DB) ([]Watchlist, error)
func DeleteWatchlist(db *sql.DB, id int64) error
func AddWatchlistItem(db *sql.DB, watchlistID int64, symbol, market string) error
func RemoveWatchlistItem(db *sql.DB, watchlistID int64, symbol string) error
func ListWatchlistItems(db *sql.DB, watchlistID int64) ([]WatchlistItem, error)
func AllWatchedSymbols(db *sql.DB) ([]string, error)
```

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/store/watchlist.go internal/store/watchlist_test.go
git commit -m "feat(store): watchlist CRUD with cascade delete"
```

---

### Task 2.3: Tasks Store

**Files:**
- Create: `internal/store/tasks.go`
- Create: `internal/store/tasks_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestCreateTask(t *testing.T)
func TestListTasks(t *testing.T)              // returns ordered by created_at
func TestToggleTask(t *testing.T)             // todo → done, done → todo
func TestDeleteTask(t *testing.T)
func TestTaskStats(t *testing.T)              // returns completed/total counts
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement tasks.go**

```go
type Task struct { ID int64; Title string; Status string; CreatedAt time.Time; CompletedAt *time.Time }

func CreateTask(db *sql.DB, title string) (int64, error)
func ListTasks(db *sql.DB) ([]Task, error)
func ToggleTask(db *sql.DB, id int64) error         // flips todo↔done, sets completed_at
func DeleteTask(db *sql.DB, id int64) error
func TaskStats(db *sql.DB) (completed, total int, err error)
```

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/store/tasks.go internal/store/tasks_test.go
git commit -m "feat(store): simple tasks CRUD with toggle"
```

---

### Task 2.4: Standup Store

**Files:**
- Create: `internal/store/standup.go`
- Create: `internal/store/standup_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestUpsertStandup(t *testing.T)          // create new, then update same date
func TestGetTodayStandup(t *testing.T)        // returns today's entry or nil
func TestListStandupHistory(t *testing.T)     // returns last N entries ordered by date desc
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement standup.go**

```go
type StandupEntry struct { ID int64; Date string; Yesterday string; Today string; Blockers string; CreatedAt time.Time; UpdatedAt *time.Time }

func UpsertStandup(db *sql.DB, date, yesterday, today, blockers string) error  // INSERT OR REPLACE
func GetStandup(db *sql.DB, date string) (*StandupEntry, error)
func GetTodayStandup(db *sql.DB) (*StandupEntry, error)
func ListStandupHistory(db *sql.DB, limit int) ([]StandupEntry, error)
```

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/store/standup.go internal/store/standup_test.go
git commit -m "feat(store): standup entries with upsert and history"
```

---

### Task 2.5: Market Cache Store

**Files:**
- Create: `internal/store/cache.go`
- Create: `internal/store/cache_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestSetAndGetQuote(t *testing.T)         // store quote, retrieve it
func TestGetQuoteStale(t *testing.T)          // quote older than TTL returns stale=true
func TestBulkSetQuotes(t *testing.T)          // batch upsert multiple quotes
func TestSetAndGetCurrencyRate(t *testing.T)  // store and retrieve rate
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement cache.go**

```go
type CachedQuote struct {
    Symbol     string; Market string
    Price      float64; Change float64; ChangePct float64
    Volume     float64; High24h float64; Low24h float64
    LastUpdated time.Time
}

func SetQuote(db *sql.DB, q CachedQuote) error
func SetQuotes(db *sql.DB, quotes []CachedQuote) error      // bulk upsert in a transaction
func GetQuote(db *sql.DB, symbol string) (*CachedQuote, error)
func GetQuotes(db *sql.DB, symbols []string) ([]CachedQuote, error)
func IsStale(q *CachedQuote, maxAge time.Duration) bool

func SetCurrencyRate(db *sql.DB, from, to string, rate float64) error
func GetCurrencyRate(db *sql.DB, from, to string) (float64, error)
```

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/store/cache.go internal/store/cache_test.go
git commit -m "feat(store): market cache and currency rate storage with staleness check"
```

---

**Phase 2 complete.** Full data layer with tests. Zero UI code — all pure business logic.

---

## Chunk 3: Phase 3 — App Shell & Layout

**Goal:** Multi-panel layout with header, footer, sidebar toggle, view switching, focus management, and input mode state machine. No real data yet — placeholder content.

### Task 3.1: Lip Gloss Theme and Keybindings

**Files:**
- Create: `internal/ui/common/styles.go`
- Create: `internal/ui/common/keys.go`

- [ ] **Step 1: Define color palette and styles in styles.go**

Use Lip Gloss v2 adaptive colors (work across light/dark terminals). Define styles for: borders, headers, focused/unfocused panels, green (gain), red (loss), muted text, table headers.

- [ ] **Step 2: Define keybindings in keys.go**

Use `bubbles/v2/key` package. Define `KeyMap` struct with bindings for: quit, view switching (1-4), sidebar toggle (Tab), add (a), delete (d), edit (e), search (/), help (?), navigation (j/k/up/down), enter, escape.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/common/
git commit -m "feat(ui): Lip Gloss theme and keybinding definitions"
```

---

### Task 3.2: Header Component

**Files:**
- Create: `internal/ui/header/model.go`

- [ ] **Step 1: Implement header model**

Renders a single-line header bar:
```
HOARD  Portfolio: $48,291  Day: +$342 (+0.71%)  AI: Groq ●  14:32
```

The header is a stateless view function that takes: total value, day change, AI provider status, current time, and terminal width. Returns a styled string.

- [ ] **Step 2: Commit**

```bash
git add internal/ui/header/
git commit -m "feat(ui): header bar component"
```

---

### Task 3.3: Footer Component

**Files:**
- Create: `internal/ui/footer/model.go`

- [ ] **Step 1: Implement footer model**

Renders bottom bar with context-sensitive key hints + error messages:
```
[1]Portfolio [2]Watchlist [3]Charts [4]News  [Tab]Sidebar  [a]dd  [?]Help  [q]Quit
```

Shows error messages when `lastErr` is set:
```
[!] Finnhub: rate limited, using cached data (2m old)
```

- [ ] **Step 2: Commit**

```bash
git add internal/ui/footer/
git commit -m "feat(ui): footer bar with key hints and error display"
```

---

### Task 3.4: Sidebar Container

**Files:**
- Create: `internal/ui/sidebar/model.go`

- [ ] **Step 1: Implement sidebar container model**

The sidebar container manages three placeholder panels (calendar, tasks, standup), stacked vertically. It handles:
- Toggle visibility
- Focus cycling between panels (j/k when sidebar is focused)
- Rendering with fixed width (~28 columns)
- Border styling (highlighted when focused)

For now, each panel renders placeholder text: "Calendar", "Tasks", "Standup".

- [ ] **Step 2: Commit**

```bash
git add internal/ui/sidebar/model.go
git commit -m "feat(ui): sidebar container with placeholder panels and focus cycling"
```

---

### Task 3.5: Root Layout Composition

**Files:**
- Modify: `internal/app/model.go`
- Create: `internal/app/update.go`
- Create: `internal/app/view.go`
- Create: `internal/app/app_test.go`

- [ ] **Step 1: Write failing tests for view switching and input modes**

```go
func TestViewSwitching(t *testing.T)       // press 1-4, verify activeView changes
func TestSidebarToggle(t *testing.T)       // press Tab, verify sidebarOpen toggles
func TestInputModeBlocking(t *testing.T)   // in modeTextInput, "1" doesn't switch view
func TestEscapeResetsMode(t *testing.T)    // Escape returns to modeNormal
func TestQuit(t *testing.T)               // "q" in normal mode returns tea.Quit
func TestQuitBlockedInInput(t *testing.T) // "q" in text input mode does NOT quit
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Refactor model.go — split Update into update.go, View into view.go**

`update.go`: Full input mode state machine. Route messages to header, market area, or sidebar based on focus. Forward WindowSizeMsg to all sub-models.

`view.go`: Compose the layout:
```
┌─ header (1 line) ─────────────────────────────────────┐
├─ market area (fills remaining) ──┬─ sidebar (28 cols) ─┤
│  [placeholder view content]      │  [sidebar panels]   │
├─ footer (1 line) ────────────────┴─────────────────────┤
```

Use `lipgloss.JoinHorizontal` for market+sidebar, `lipgloss.JoinVertical` for header+body+footer.

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Build and manually verify layout**

```bash
make run
# Expected: header bar, content area with view label, sidebar (if open), footer with key hints
# Press 1-4: view label changes. Tab: sidebar toggles. q: quits.
```

- [ ] **Step 6: Commit**

```bash
git add internal/app/
git commit -m "feat(app): multi-panel layout with sidebar toggle, view switching, input mode state machine"
```

---

### Task 3.6: Placeholder Market Views

**Files:**
- Create: `internal/ui/market/portfolio/model.go`
- Create: `internal/ui/market/portfolio/view.go`
- Create: `internal/ui/market/watchlist/model.go`
- Create: `internal/ui/market/watchlist/view.go`
- Create: `internal/ui/market/charts/model.go`
- Create: `internal/ui/market/charts/view.go`
- Create: `internal/ui/market/news/model.go`
- Create: `internal/ui/market/news/view.go`

- [ ] **Step 1: Create stub models for all four views**

Each view model implements:
```go
type PortfolioModel struct { width, height int }
func New() PortfolioModel { ... }
func (m PortfolioModel) Init() tea.Cmd { return nil }
func (m PortfolioModel) Update(msg tea.Msg) (PortfolioModel, tea.Cmd) { ... }
func (m PortfolioModel) View() string { return "Portfolio view — coming soon" }
func (m *PortfolioModel) SetSize(w, h int) { m.width = w; m.height = h }
```

Same pattern for Watchlist, Charts, News — just different placeholder text.

- [ ] **Step 2: Wire stub views into the root App model**

In `app/model.go`, add fields for each view model. In `app/view.go`, render the active view's `View()` in the market area.

- [ ] **Step 3: Build and manually test all view switches**

```bash
make run
# Press 1: "Portfolio view", 2: "Watchlist view", 3: "Charts view", 4: "News view"
```

- [ ] **Step 4: Commit**

```bash
git add internal/ui/market/
git commit -m "feat(ui): stub models for portfolio, watchlist, charts, and news views"
```

---

**Phase 3 complete.** You have a fully navigable TUI shell: header, footer, four switchable views, collapsible sidebar, input mode state machine. No real data yet — that's next.

---

## Chunk 4: Phase 4 — Portfolio View & Market Data

**Goal:** Working portfolio view with real stock/crypto prices. The core value of the app.

### Task 4.1: MarketProvider Interface and Finnhub Client

**Files:**
- Create: `internal/service/market/provider.go`
- Create: `internal/service/market/finnhub.go`
- Create: `internal/service/market/finnhub_test.go`

- [ ] **Step 1: Define MarketProvider interface**

```go
// internal/service/market/provider.go
package market

import "context"

type Quote struct {
    Symbol    string
    Price     float64
    Change    float64
    ChangePct float64
    Volume    float64
    High      float64
    Low       float64
}

type Candle struct {
    Time   int64
    Open   float64
    High   float64
    Low    float64
    Close  float64
    Volume float64
}

type Article struct {
    Headline  string
    Summary   string
    Source    string
    URL       string
    Datetime  int64
    Symbol    string
    Sentiment float64 // -1 to 1, 0 = neutral
}

type StockProvider interface {
    GetQuote(ctx context.Context, symbol string) (*Quote, error)
    GetQuotes(ctx context.Context, symbols []string) ([]Quote, error)
    GetCandles(ctx context.Context, symbol string, from, to int64, resolution string) ([]Candle, error)
    SearchSymbol(ctx context.Context, query string) ([]SymbolMatch, error)
    GetNews(ctx context.Context, symbol string) ([]Article, error)
}

type CryptoProvider interface {
    GetQuote(ctx context.Context, symbol string) (*Quote, error)
    GetQuotes(ctx context.Context, symbols []string) ([]Quote, error)
    GetCandles(ctx context.Context, symbol string, days int) ([]Candle, error)
}

type SymbolMatch struct {
    Symbol      string
    Description string
    Type        string // "Common Stock", "ETF", "Crypto"
}
```

- [ ] **Step 2: Write failing test for Finnhub client (using HTTP mock)**

```go
// internal/service/market/finnhub_test.go
func TestFinnhubGetQuote(t *testing.T)       // mock /quote endpoint, verify parsing
func TestFinnhubGetQuotes(t *testing.T)      // batch, verify all returned
func TestFinnhubSearchSymbol(t *testing.T)   // mock /search, verify results
func TestFinnhubTimeout(t *testing.T)        // verify context timeout is respected
```

Use `httptest.NewServer` to mock Finnhub responses.

- [ ] **Step 3: Run tests — verify they fail**
- [ ] **Step 4: Implement finnhub.go**

Use the official `finnhub-go/v2` SDK. Wrap it with our interface. Add `context.WithTimeout`.

- [ ] **Step 5: Run tests — verify they pass**
- [ ] **Step 6: Commit**

```bash
git add internal/service/market/
git commit -m "feat(service): MarketProvider interface and Finnhub client with mock tests"
```

---

### Task 4.2: CoinGecko Client

**Files:**
- Create: `internal/service/market/coingecko.go`
- Create: `internal/service/market/coingecko_test.go`

- [ ] **Step 1: Write failing tests (HTTP mock)**
- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement coingecko.go using `goingecko` library**

Handle symbol mapping: our format `BTC-USD` → CoinGecko ID `bitcoin`. Maintain a small mapping table for top coins; fall back to search API for unknown symbols.

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/service/market/coingecko.go internal/service/market/coingecko_test.go
git commit -m "feat(service): CoinGecko client with symbol mapping"
```

---

### Task 4.3: Market Cache Service Layer

**Files:**
- Create: `internal/service/market/cache.go`

- [ ] **Step 1: Implement caching wrapper**

Wraps a `StockProvider` and `CryptoProvider`. On `GetQuotes`:
1. Check SQLite cache — if fresh (< TTL), return cached
2. If stale, call underlying provider
3. Store result in cache
4. On provider error, return stale cache with a warning flag

```go
type CachedMarketService struct {
    stocks  StockProvider
    crypto  CryptoProvider
    db      *sql.DB
    stockTTL time.Duration
    cryptoTTL time.Duration
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/service/market/cache.go
git commit -m "feat(service): market cache layer with TTL and stale fallback"
```

---

### Task 4.4: Currency Service

**Files:**
- Create: `internal/service/currency/frankfurter.go`
- Create: `internal/service/currency/frankfurter_test.go`

- [ ] **Step 1: Write failing test**

Mock the Frankfurter API response. Verify parsing of `{"base":"USD","date":"2026-03-13","rates":{"EUR":0.92,"GBP":0.79}}`.

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement frankfurter.go**

Simple HTTP GET to `https://api.frankfurter.dev/v1/latest?base=USD`. Parse response, store rates in SQLite. Check cache first — only fetch if last fetch > 24h.

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/service/currency/
git commit -m "feat(service): Frankfurter currency rate fetcher with daily cache"
```

---

### Task 4.5: Portfolio View — Holdings Table

**Files:**
- Modify: `internal/ui/market/portfolio/model.go`
- Modify: `internal/ui/market/portfolio/view.go`
- Create: `internal/ui/market/portfolio/model_test.go`

- [ ] **Step 1: Write failing test for P&L calculation**

```go
func TestCalculatePnL(t *testing.T)           // quantity * (price - avg_cost)
func TestCalculateAllocation(t *testing.T)     // holding value / total value * 100
func TestDayChangeFormatting(t *testing.T)     // positive green, negative red
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement portfolio model**

Uses `bubbles/v2/table` component for the holdings table. Columns: Symbol, Shares, Avg Cost, Price, Day Chg, P&L, Alloc %. Receives market data via messages. Supports sorting by any column.

- [ ] **Step 4: Implement portfolio view**

Render: summary line (total value, day change, total P&L) + table + keybind hints.

- [ ] **Step 5: Run tests — verify they pass**
- [ ] **Step 6: Commit**

```bash
git add internal/ui/market/portfolio/
git commit -m "feat(ui): portfolio view with holdings table, P&L, and allocation"
```

---

### Task 4.6: Add Position Form

**Files:**
- Modify: `internal/ui/market/portfolio/model.go`

- [ ] **Step 1: Implement inline add/edit form**

When user presses `a` in portfolio view:
1. Switch to `modeTextInput`
2. Show form: Symbol → Quantity → Avg Cost → Account → Market (auto-detected from symbol)
3. Use `bubbles/v2/textinput` for each field
4. Tab/Enter cycles between fields
5. Submit saves to DB via store functions
6. Escape cancels, returns to modeNormal

- [ ] **Step 2: Test form transitions**

```go
func TestAddFormOpens(t *testing.T)        // press 'a', mode switches to text input
func TestAddFormSubmit(t *testing.T)       // fill fields, submit, holding created
func TestAddFormCancel(t *testing.T)       // press Escape, form closes, no data saved
```

- [ ] **Step 3: Run tests — verify they pass**
- [ ] **Step 4: Commit**

```bash
git add internal/ui/market/portfolio/
git commit -m "feat(ui): add position form with inline text inputs"
```

---

### Task 4.7: Wire Market Data into TUI Tick Loop

**Files:**
- Modify: `internal/app/model.go`
- Modify: `internal/app/update.go`
- Modify: `internal/app/messages.go`

- [ ] **Step 1: Add market service to App model**

```go
type App struct {
    // ... existing fields
    marketSvc *market.CachedMarketService
    currSvc   *currency.FrankfurterService
}
```

- [ ] **Step 2: Implement tick-based refresh in update.go**

In `Init()`: start first tick + initial data fetch.
In `Update()`: handle `TickMsg`, `QuotesMsg`, `CryptoQuotesMsg`, `CurrencyRatesMsg`, `ErrMsg`.

Follow the pattern from architecture doc Section 4 — adaptive polling based on market hours.

- [ ] **Step 3: Wire market data into portfolio view**

`QuotesMsg` → update portfolio model's price data → recalculate P&L → re-render table.

- [ ] **Step 4: Build and test with real API key**

```bash
HOARD_FINNHUB_KEY=your_key make run
# Expected: portfolio shows real-time prices (if you've added holdings)
```

- [ ] **Step 5: Commit**

```bash
git add internal/app/
git commit -m "feat: wire market data into TUI tick loop with adaptive polling"
```

---

**Phase 4 complete.** You now have a working portfolio view with real stock/crypto prices, add position form, P&L calculations, and adaptive market-hours-aware polling.

---

## Chunk 5: Phase 5 — Watchlist & Charts

**Goal:** Watchlist management and price charts with timeframe selection.

### Task 5.1: Watchlist View

**Files:**
- Modify: `internal/ui/market/watchlist/model.go`
- Modify: `internal/ui/market/watchlist/view.go`
- Create: `internal/ui/market/watchlist/model_test.go`

- [ ] **Step 1: Write failing tests for watchlist model**

```go
func TestCreateWatchlistUI(t *testing.T)      // new watchlist via 'n' key
func TestAddSymbolToWatchlist(t *testing.T)   // add via 'a' key
func TestRemoveSymbolUI(t *testing.T)         // remove via 'd' key
func TestMoveToPortfolioUI(t *testing.T)      // 'm' triggers add position form pre-filled
func TestSwitchWatchlist(t *testing.T)        // cycle between named watchlists
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement watchlist model and view**

Display: multiple named watchlists as columns (or tabs), each showing symbol + price + day change. Live prices from the same market data tick loop. Symbol search with `bubbles/v2/textinput` and fuzzy matching.

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/ui/market/watchlist/
git commit -m "feat(ui): watchlist view with named lists, add/remove, and move-to-portfolio"
```

---

### Task 5.2: ChartRenderer Interface

**Files:**
- Create: `internal/ui/common/chart.go`

- [ ] **Step 1: Define ChartRenderer interface**

```go
// internal/ui/common/chart.go
package common

type ChartData struct {
    Labels []string  // x-axis labels (dates/times)
    Values []float64 // y-axis values (prices)
}

type ChartRenderer interface {
    RenderLineChart(data ChartData, width, height int) string
    RenderSparkline(values []float64, width int) string
}
```

- [ ] **Step 2: Implement a basic renderer using asciigraph**

This is the safe fallback. If ntcharts gets v2 support, swap the implementation later.

```go
type AsciigraphRenderer struct{}

func (r AsciigraphRenderer) RenderLineChart(data ChartData, width, height int) string {
    return asciigraph.Plot(data.Values, asciigraph.Width(width), asciigraph.Height(height))
}

func (r AsciigraphRenderer) RenderSparkline(values []float64, width int) string {
    // Braille sparkline implementation (~50 lines)
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/ui/common/chart.go
git commit -m "feat(ui): ChartRenderer interface with asciigraph fallback"
```

---

### Task 5.3: Charts View — Price History

**Files:**
- Modify: `internal/ui/market/charts/model.go`
- Modify: `internal/ui/market/charts/view.go`
- Create: `internal/ui/market/charts/model_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestTimeframeSwitch(t *testing.T)       // keys cycle through 1D/1W/1M/3M/1Y/ALL
func TestSymbolSelection(t *testing.T)       // selecting a holding shows its chart
func TestIndicatorToggle(t *testing.T)       // MA/RSI toggle on/off
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement charts model**

Model holds: selected symbol, timeframe, candle data, indicator toggles. On symbol/timeframe change, fetches candles from market service (via command). View renders chart using ChartRenderer + indicator overlays + info line (MA values, RSI, volume).

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/ui/market/charts/
git commit -m "feat(ui): charts view with timeframe selection and technical indicators"
```

---

### Task 5.4: MA and RSI Calculations

**Files:**
- Create: `internal/ui/market/charts/indicators.go`
- Create: `internal/ui/market/charts/indicators_test.go`

- [ ] **Step 1: Write failing tests for indicator math**

```go
func TestSMA(t *testing.T)      // simple moving average over N periods
func TestEMA(t *testing.T)      // exponential moving average
func TestRSI(t *testing.T)      // relative strength index (14 periods)
```

Use known test vectors (hand-calculated or from a finance reference).

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement pure functions**

```go
func SMA(prices []float64, period int) []float64
func EMA(prices []float64, period int) []float64
func RSI(prices []float64, period int) []float64
```

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/ui/market/charts/indicators.go internal/ui/market/charts/indicators_test.go
git commit -m "feat: SMA, EMA, RSI indicator calculations with test vectors"
```

---

### Task 5.5: Wire Charts to Market Service

**Files:**
- Modify: `internal/app/update.go`
- Modify: `internal/app/messages.go`

- [ ] **Step 1: Add CandlesMsg message type**
- [ ] **Step 2: When charts view is active and symbol/timeframe changes, dispatch FetchCandles command**
- [ ] **Step 3: Handle CandlesMsg — update charts model with candle data**
- [ ] **Step 4: Build and test with real data**

```bash
HOARD_FINNHUB_KEY=your_key make run
# Add a holding (AAPL), switch to Charts view [3], verify chart renders
```

- [ ] **Step 5: Commit**

```bash
git add internal/app/
git commit -m "feat: wire charts view to market service with candle data fetching"
```

---

**Phase 5 complete.** Watchlist management + price charts with MA/RSI indicators.

---

## Chunk 6: Phase 6 — News & Sidebar

**Goal:** News feed, calendar agenda, task list, and standup log — all functional.

### Task 6.1: News View

**Files:**
- Modify: `internal/ui/market/news/model.go`, `view.go`
- Create: `internal/ui/market/news/model_test.go`

- [ ] **Step 1: Write failing tests** — news filtering by held symbols, keyword sentiment scoring
- [ ] **Step 2: Implement news model** — fetches articles via Finnhub news endpoint, filters to portfolio+watchlist symbols, scores sentiment via keyword heuristic (positive: "beat", "surge", "record"; negative: "miss", "crash", "recall")
- [ ] **Step 3: Implement news view** — list of articles with sentiment dot, source, time. `Enter` to expand, `o` to open in browser (`exec.Command("open", url)`)
- [ ] **Step 4: Wire to tick loop** — refresh every 5 minutes
- [ ] **Step 5: Commit**

```bash
git add internal/ui/market/news/
git commit -m "feat(ui): news view with symbol filtering and keyword sentiment"
```

---

### Task 6.2: Calendar Provider and ICS Parser

**Files:**
- Create: `internal/service/calendar/provider.go`
- Create: `internal/service/calendar/ics.go`
- Create: `internal/service/calendar/ics_test.go`

- [ ] **Step 1: Define CalendarProvider interface**

```go
type Event struct { ID string; Title string; Start time.Time; End time.Time; Calendar string }
type CalendarProvider interface { FetchEvents(ctx context.Context, from, to time.Time) ([]Event, error) }
```

- [ ] **Step 2: Write failing ICS parser tests** — use a sample `.ics` file embedded in test
- [ ] **Step 3: Implement ics.go** — parse VEVENT blocks, extract SUMMARY, DTSTART, DTEND
- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/service/calendar/
git commit -m "feat(service): CalendarProvider interface and ICS file parser"
```

---

### Task 6.3: Google Workspace CLI Calendar

**Files:**
- Create: `internal/service/calendar/gws.go`

- [ ] **Step 1: Implement gws.go** — shells out to `gws calendar events list --format json`, parses output. Returns error if `gws` not found in PATH (triggers fallback to next provider).

- [ ] **Step 2: Commit**

```bash
git add internal/service/calendar/gws.go
git commit -m "feat(service): Google Workspace CLI calendar integration"
```

---

### Task 6.4: Calendar Sidebar Panel

**Files:**
- Modify: `internal/ui/sidebar/calendar/model.go`

- [ ] **Step 1: Implement calendar panel** — shows today's events + tomorrow peek. Formatted as time + title, one per line. Updates via CalendarEventsMsg from tick loop (every 15 min).
- [ ] **Step 2: Commit**

```bash
git add internal/ui/sidebar/calendar/
git commit -m "feat(ui): calendar sidebar panel with today/tomorrow agenda"
```

---

### Task 6.5: Tasks Sidebar Panel

**Files:**
- Modify: `internal/ui/sidebar/tasks/model.go`
- Create: `internal/ui/sidebar/tasks/model_test.go`

- [ ] **Step 1: Write failing tests** — add task, toggle, delete, navigation
- [ ] **Step 2: Implement tasks panel** — renders task list from store. `a`: add (switches to modeTextInput), `x`: toggle done, `d`: delete, `j/k`: navigate. Header shows count (completed/total).
- [ ] **Step 3: Run tests — verify they pass**
- [ ] **Step 4: Commit**

```bash
git add internal/ui/sidebar/tasks/
git commit -m "feat(ui): tasks sidebar panel with add/toggle/delete"
```

---

### Task 6.6: Standup Sidebar Panel

**Files:**
- Modify: `internal/ui/sidebar/standup/model.go`
- Create: `internal/ui/sidebar/standup/model_test.go`

- [ ] **Step 1: Write failing tests** — edit standup, view history, git commit integration
- [ ] **Step 2: Implement standup panel** — shows Y/T/B for today. `e`: edit (switches to modeTextInput with three text areas). `h`: history view (scrollable list). `g`: generate draft from `git log --since=yesterday --oneline` (runs command, shows raw output for now — AI enhancement in Phase 7).
- [ ] **Step 3: Run tests — verify they pass**
- [ ] **Step 4: Commit**

```bash
git add internal/ui/sidebar/standup/
git commit -m "feat(ui): standup sidebar panel with editor, history, and git log draft"
```

---

**Phase 6 complete.** All four market views and all three sidebar panels are functional.

---

## Chunk 7: Phase 7 — AI Layer

**Goal:** AIProvider interface with fallback chain. AI features integrated into news summary, standup draft, and sentiment analysis.

### Task 7.1: AIProvider Interface and Noop Provider

**Files:**
- Create: `internal/service/ai/provider.go`
- Create: `internal/service/ai/noop.go`

- [ ] **Step 1: Define interface**

```go
type AIProvider interface {
    Name() string
    Available() bool  // quick check: API key set, endpoint reachable
    Complete(ctx context.Context, prompt string) (string, error)
    Summarize(ctx context.Context, text string) (string, error)
}
```

- [ ] **Step 2: Implement noop.go** — returns empty string, nil error. Always "available". This is the terminal fallback.
- [ ] **Step 3: Commit**

```bash
git add internal/service/ai/provider.go internal/service/ai/noop.go
git commit -m "feat(ai): AIProvider interface and noop fallback"
```

---

### Task 7.2: Ollama, Groq, Gemini Providers

**Files:**
- Create: `internal/service/ai/ollama.go`
- Create: `internal/service/ai/groq.go`
- Create: `internal/service/ai/gemini.go`

- [ ] **Step 1: Implement ollama.go** — POST to `{endpoint}/api/generate` (Ollama's native API). `Available()` = HTTP ping to endpoint. Uses configured model.

- [ ] **Step 2: Implement groq.go** — POST to `https://api.groq.com/openai/v1/chat/completions` (OpenAI-compatible). `Available()` = API key is set. Default model: `llama-3.3-70b-versatile`.

- [ ] **Step 3: Implement gemini.go** — POST to `https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent`. `Available()` = API key is set. Default model: `gemini-2.5-flash`.

All providers: use `context.WithTimeout`, structured error returns, response parsing.

- [ ] **Step 4: Commit**

```bash
git add internal/service/ai/ollama.go internal/service/ai/groq.go internal/service/ai/gemini.go
git commit -m "feat(ai): Ollama, Groq, and Gemini provider implementations"
```

---

### Task 7.3: Fallback Chain

**Files:**
- Create: `internal/service/ai/chain.go`
- Create: `internal/service/ai/chain_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestChainUsesFirstAvailable(t *testing.T)   // first provider works → used
func TestChainFallsBack(t *testing.T)            // first fails → second used
func TestChainAllFail(t *testing.T)              // all fail → noop (empty string, no error)
func TestChainCachesResponse(t *testing.T)       // same prompt within TTL → cached result
func TestChainReportsActiveProvider(t *testing.T) // ActiveProvider() returns name of last used
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement chain.go**

```go
type Chain struct {
    providers []AIProvider
    cache     map[string]cachedResult  // prompt hash → result + timestamp
    cacheTTL  time.Duration
    lastUsed  string
}

func NewChain(providers []AIProvider, cacheTTL time.Duration) *Chain
func (c *Chain) Complete(ctx context.Context, prompt string) (string, error)
func (c *Chain) Summarize(ctx context.Context, text string) (string, error)
func (c *Chain) ActiveProvider() string
```

Iterates providers in order. Skips unavailable. Catches errors and tries next. If all fail, returns `""` with nil error (graceful degradation — feature just doesn't enhance).

- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Commit**

```bash
git add internal/service/ai/chain.go internal/service/ai/chain_test.go
git commit -m "feat(ai): fallback chain with caching and graceful degradation"
```

---

### Task 7.4: Integrate AI into News View

**Files:**
- Modify: `internal/ui/market/news/model.go`
- Modify: `internal/app/update.go`
- Modify: `internal/app/messages.go`

- [ ] **Step 1: Add `[s]` keybind** — on selected article, dispatch AI summarize command
- [ ] **Step 2: Add AI sentiment** — when AI is available, score articles with `Complete("Rate sentiment: bullish/bearish/neutral: {headline}")`. Fall back to keyword heuristic.
- [ ] **Step 3: Add `[w]` "Why did it move?"** — on selected holding, prompt: "Explain why {symbol} moved {change}% today based on these headlines: {headlines}". Show AI response in an overlay/viewport.
- [ ] **Step 4: Commit**

```bash
git add internal/ui/market/news/ internal/app/
git commit -m "feat: AI news summarization, sentiment scoring, and 'why did it move' explanation"
```

---

### Task 7.5: Integrate AI into Standup

**Files:**
- Modify: `internal/ui/sidebar/standup/model.go`

- [ ] **Step 1: Enhance `[g]` command** — instead of raw git log, pass commits to AI: "Convert these git commits into a concise standup update with Yesterday/Today/Blockers format: {commits}". Fall back to raw commit list if AI unavailable.
- [ ] **Step 2: Show AI provider status in header** — update header to show active provider name and indicator.
- [ ] **Step 3: Commit**

```bash
git add internal/ui/sidebar/standup/ internal/ui/header/
git commit -m "feat: AI-powered standup draft generation and provider status in header"
```

---

**Phase 7 complete.** AI fallback chain with news summary, sentiment, "why did it move", and standup generation.

---

## Chunk 8: Phase 8 — Polish & Release

**Goal:** First-run experience, CSV import, error display, GoReleaser, README.

### Task 8.1: First-Run Experience

**Files:**
- Modify: `internal/app/view.go`
- Modify: `internal/ui/market/portfolio/view.go`

- [ ] **Step 1: Empty state views** — when portfolio is empty, show: "Welcome to Hoard! Press [a] to add your first position." Same pattern for empty watchlists, empty tasks.
- [ ] **Step 2: API key prompting** — when first market data fetch fails due to missing key, show inline prompt: "Enter your Finnhub API key (free at finnhub.io):" → save to config.
- [ ] **Step 3: Auto-create config** — on first run, write `config.example.toml` contents to `~/.config/hoard/config.toml` with comments.
- [ ] **Step 4: Commit**

```bash
git add internal/
git commit -m "feat: first-run experience with empty states and API key prompting"
```

---

### Task 8.2: Robinhood CSV Import

**Files:**
- Create: `internal/importer/importer.go`
- Create: `internal/importer/robinhood.go`
- Create: `internal/importer/robinhood_test.go`

- [ ] **Step 1: Write failing tests with sample Robinhood CSV**

Robinhood CSV format (Activity → Account Statements → Download):
```csv
Activity Date,Process Date,Settle Date,Instrument,Description,Trans Code,Quantity,Price,Amount
03/10/2026,03/10/2026,03/12/2026,AAPL,APPLE INC,Buy,10,198.42,-1984.20
03/05/2026,03/05/2026,03/07/2026,GOOGL,ALPHABET INC,Buy,5,171.20,-856.00
```

```go
func TestParseRobinhoodCSV(t *testing.T)         // parse sample CSV, verify transactions
func TestRobinhoodCSVBadFormat(t *testing.T)      // invalid CSV returns clear error
func TestImportToStore(t *testing.T)              // parsed transactions saved to DB
```

- [ ] **Step 2: Run tests — verify they fail**
- [ ] **Step 3: Implement robinhood.go** — parse CSV, map to Transaction structs, handle Buy/Sell/Dividend trans codes
- [ ] **Step 4: Run tests — verify they pass**
- [ ] **Step 5: Add import command to TUI** — `hoard import robinhood path/to/file.csv` (CLI subcommand) or in-app `[i]mport` key
- [ ] **Step 6: Commit**

```bash
git add internal/importer/
git commit -m "feat: Robinhood CSV import with parser and store integration"
```

---

### Task 8.3: Error Display Polish

**Files:**
- Modify: `internal/ui/footer/model.go`
- Modify: `internal/app/update.go`

- [ ] **Step 1: Implement timed error messages** — `ErrMsg` shows in footer for 10 seconds, then auto-clears. Use `tea.Tick` for auto-dismiss.
- [ ] **Step 2: Implement error severity levels** — warning (yellow) vs error (red). Rate limit warnings are yellow; connection failures are red.
- [ ] **Step 3: Commit**

```bash
git add internal/ui/footer/ internal/app/
git commit -m "feat: timed error display in footer with severity levels"
```

---

### Task 8.4: GoReleaser Configuration

**Files:**
- Create: `.goreleaser.yaml`
- Modify: `Makefile` (add release target)

- [ ] **Step 1: Write .goreleaser.yaml**

```yaml
version: 2
builds:
  - main: ./cmd/hoard
    binary: hoard
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - formats: ['tar.gz']
    format_overrides:
      - goos: windows
        formats: ['zip']
    name_template: "hoard_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

brews:
  - repository:
      owner: yourusername
      name: homebrew-tap
    homepage: "https://github.com/yourusername/hoard"
    description: "Terminal finance dashboard & daily cockpit"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^chore:"
      - "^test:"
```

- [ ] **Step 2: Test with `goreleaser check`**

```bash
goreleaser check
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yaml Makefile
git commit -m "chore: GoReleaser config for cross-platform releases + Homebrew"
```

---

### Task 8.5: README with Demo

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README** with:
- Project name + one-line description
- Demo GIF placeholder (record with VHS later)
- Features list
- Installation (go install, Homebrew, binary download)
- Quick start guide
- Configuration reference
- Keybindings reference
- Recommended fonts
- License (MIT)

- [ ] **Step 2: Record demo GIF with VHS**

Install VHS: `go install github.com/charmbracelet/vhs@latest`

Create `demo.tape` with scripted demo: launch → show portfolio → switch views → toggle sidebar → add position → show chart.

```bash
vhs demo.tape
```

- [ ] **Step 3: Commit**

```bash
git add README.md demo.tape demo.gif
git commit -m "docs: README with features, installation, keybindings, and demo GIF"
```

---

**Phase 8 complete. MVP is done.**

---

## Post-MVP Roadmap (Not Planned in Detail)

These features are scoped in `docs/design-spec.md` Section 8 but not broken into implementation tasks yet:

1. **Deep Research via notebooklm-py** — Stock research mode with grounded Q&A
2. **Gmail via gws** — Parse trade confirmation emails
3. **Google Sheets import** — Portfolio from spreadsheet
4. **Post/Share** — Screenshot portfolio to social
5. **International markets** — Non-US exchanges
6. **WebSocket streaming** — Real-time prices via Finnhub WebSocket
7. **Advanced charting** — More indicators, drawing tools
8. **Brokerage aggregation** — SnapTrade/Plaid integration

Each should follow the same pattern: spec → plan → implement → test → release.
