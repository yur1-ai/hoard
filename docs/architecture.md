# Hoard — Technical Architecture

**Date:** 2026-03-13
**Stack:** Go + Bubble Tea v2 + SQLite (modernc) + Lip Gloss v2

---

## 1. Bubble Tea v2 Component Tree

```
App (root model)
│   Owns: layout state, focus tracking, window size, input mode, global keybinds
│
├── Header
│   "HOARD  Portfolio: $48,291  Day: +$342 (+0.71%)  AI: Groq ●  14:32"
│
├── Market Area ─────────────────────────── Sidebar (collapsible)
│   │                                       │
│   ├── PortfolioModel [1]                  ├── CalendarModel
│   │   Holdings table, P&L, allocation     │   Today/tomorrow agenda
│   │                                       │
│   ├── WatchlistModel [2]                  ├── TasksModel
│   │   Named lists, quick-add              │   Simple todo CRUD
│   │                                       │
│   ├── ChartsModel   [3]                   └── StandupModel
│   │   Price chart, indicators, timeframe      Y/T/B editor, history
│   │
│   └── NewsModel     [4]
│       Filtered headlines, AI summary
│
└── Footer
    "[1-4] views  [Tab] sidebar  [/] search  [a]dd  [?] help  [q]uit"
```

---

## 2. Message Flow

```
                    ┌──────────────┐
                    │   tea.Msg    │  ← user input, API responses, ticks
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │  App.Update  │  ← routes msg based on focus + msg type
                    └──┬───┬───┬──┘
                       │   │   │
          ┌────────────┘   │   └────────────┐
          ▼                ▼                ▼
   MarketArea.Update  Sidebar.Update   Global handlers
          │                │           (resize, quit, tab)
          ▼                ▼
   ActiveView.Update  ActivePanel.Update
          │                │
          ▼                ▼
     (model, cmd)     (model, cmd)
          │                │
          └───────┬────────┘
                  ▼
            tea.Batch(cmds...)  → spawns async work → returns new Msgs
```

### Key v2 Patterns
- `View()` returns `tea.View` struct (not `string`) with declarative fields
- Import paths: `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2`
- `tea.KeyPressMsg` / `tea.KeyReleaseMsg` (replaces old `tea.KeyMsg`)
- `key.Code` + `key.Text` + `key.Mod` (replaces `key.Type` + `key.Runes`)
- Color auto-downsampling: TrueColor → ANSI256 → ANSI → ASCII (built-in)

---

## 3. Input Mode State Machine

**Critical: must be implemented from day one to prevent shortcut conflicts.**

```go
type inputMode int
const (
    modeNormal    inputMode = iota  // single-key shortcuts active
    modeTextInput                    // all keys forwarded to active input
    modeSearch                       // fuzzy search overlay
)

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        // Escape ALWAYS returns to normal mode
        if msg.Code == "escape" {
            m.mode = modeNormal
            m.activeInput.Blur()
            return m, nil
        }

        switch m.mode {
        case modeNormal:
            return m.handleNormalKeys(msg)
        case modeTextInput:
            // Forward ALL keys to active input — no shortcut interception
            return m.forwardToInput(msg)
        case modeSearch:
            return m.handleSearchKeys(msg)
        }
    }
    // ... other message types
}
```

---

## 4. Data Refresh Strategy

```go
func (m App) scheduleTick() tea.Cmd {
    interval := 60 * time.Second  // default

    if isUSMarketOpen() {
        interval = 30 * time.Second  // stocks: 30s during trading hours
    }

    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

// On each tick, batch all needed refreshes
case TickMsg:
    cmds := []tea.Cmd{m.scheduleTick()}

    if m.market.NeedsRefresh("equity") {
        cmds = append(cmds, m.services.Market.FetchQuotes(m.store.AllEquitySymbols()))
    }
    if m.market.NeedsRefresh("crypto") {
        cmds = append(cmds, m.services.Market.FetchCrypto(m.store.AllCryptoSymbols()))
    }
    if m.sidebar.Visible && m.calendar.Stale(15*time.Minute) {
        cmds = append(cmds, m.services.Calendar.FetchEvents())
    }
    return m, tea.Batch(cmds...)
```

**Market hours awareness:**
- US market: 9:30-16:00 ET (Mon-Fri, excl. holidays)
- Pre/post-market: 4:00-9:30, 16:00-20:00 ET (reduced polling)
- Crypto: 24/7, always poll
- Off-hours: 5 min interval (price can still move in after-hours)

---

## 5. Project Structure

```
hoard/
├── cmd/
│   └── hoard/
│       └── main.go                 # Entry point, config loading, tea.NewProgram
│
├── internal/
│   ├── app/
│   │   ├── model.go                # Root App model, Init(), layout state
│   │   ├── update.go               # Root Update — message routing + global keys
│   │   ├── view.go                 # Root View — composes header/market/sidebar/footer
│   │   └── messages.go             # All message types in one place
│   │
│   ├── ui/
│   │   ├── market/
│   │   │   ├── portfolio/
│   │   │   │   ├── model.go        # Holdings table, P&L calculations
│   │   │   │   └── view.go         # Table rendering with sparklines
│   │   │   ├── watchlist/
│   │   │   │   ├── model.go        # Multiple named lists
│   │   │   │   └── view.go
│   │   │   ├── charts/
│   │   │   │   ├── model.go        # Chart data, timeframe, indicators
│   │   │   │   └── view.go         # Braille/block chart rendering
│   │   │   └── news/
│   │   │       ├── model.go        # Filtered articles, AI summary state
│   │   │       └── view.go
│   │   │
│   │   ├── sidebar/
│   │   │   ├── model.go            # Sidebar container (toggle, focus cycling)
│   │   │   ├── calendar/
│   │   │   │   └── model.go        # Agenda display
│   │   │   ├── tasks/
│   │   │   │   └── model.go        # CRUD, toggle, navigate
│   │   │   └── standup/
│   │   │       └── model.go        # Y/T/B editor, history
│   │   │
│   │   └── common/
│   │       ├── styles.go           # Lip Gloss theme (colors, borders, spacing)
│   │       ├── keys.go             # Keybinding definitions (help.KeyMap)
│   │       ├── table.go            # Reusable styled table component
│   │       └── chart.go            # Chart rendering utilities / abstraction
│   │
│   ├── service/
│   │   ├── market/
│   │   │   ├── provider.go         # MarketProvider interface
│   │   │   ├── finnhub.go          # Finnhub implementation (stocks/ETFs)
│   │   │   ├── coingecko.go        # CoinGecko implementation (crypto)
│   │   │   └── cache.go            # Rate-limit-aware caching layer
│   │   ├── calendar/
│   │   │   ├── provider.go         # CalendarProvider interface
│   │   │   ├── gws.go              # Google Workspace CLI (shell out)
│   │   │   ├── google.go           # Direct Google Calendar API
│   │   │   └── ics.go              # Local .ics file parser
│   │   ├── ai/
│   │   │   ├── provider.go         # AIProvider interface
│   │   │   ├── chain.go            # Fallback chain (groq → gemini → ollama → noop)
│   │   │   ├── groq.go
│   │   │   ├── gemini.go
│   │   │   ├── ollama.go
│   │   │   └── noop.go             # No-AI fallback (returns empty, never errors)
│   │   └── currency/
│   │       └── frankfurter.go      # Currency rates (daily refresh)
│   │
│   ├── store/
│   │   ├── db.go                   # SQLite init, migrations runner, WAL management
│   │   ├── portfolio.go            # Holdings + transactions CRUD
│   │   ├── watchlist.go            # Watchlist CRUD
│   │   ├── tasks.go                # Tasks CRUD
│   │   ├── standup.go              # Standup entries CRUD
│   │   └── cache.go                # Market cache read/write
│   │
│   ├── config/
│   │   ├── config.go               # TOML parsing, defaults, validation
│   │   └── paths.go                # XDG-aware path resolution
│   │
│   └── importer/
│       ├── importer.go             # Importer interface
│       └── robinhood.go            # Robinhood CSV parser
│
├── migrations/
│   ├── 001_initial_schema.sql
│   └── embed.go                    # go:embed for migration files
│
├── docs/
│   ├── design-spec.md
│   └── architecture.md
│
├── config.example.toml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 6. Key Dependencies

```go
// go.mod
module github.com/yourusername/hoard

go 1.23

require (
    // TUI framework (v2)
    charm.land/bubbletea   v2.x
    charm.land/lipgloss    v2.x
    charm.land/bubbles     v2.x
    charm.land/glamour     v0.x  // Markdown rendering (news articles)

    // Data
    modernc.org/sqlite     v1.x  // Pure Go SQLite — no CGO

    // Config
    github.com/pelletier/go-toml/v2  v2.x

    // Market data
    github.com/Finnhub-Stock-API/finnhub-go/v2  v2.x  // Stocks/ETFs
    github.com/JulianToledano/goingecko         v3.x  // Crypto (CoinGecko)

    // Charts — NOTE: check v2 compatibility before depending
    github.com/NimbleMarkets/ntcharts  v0.x  // Bubble Tea charts
    // FALLBACK: github.com/guptarohit/asciigraph (zero-dep string renderer)
)
```

### Dependency Notes

| Dependency | Note |
|-----------|------|
| `modernc.org/sqlite` | Pure Go, no CGO, easy cross-compile. Pin `modernc.org/libc` version exactly. |
| `ntcharts` | Currently imports Bubble Tea v1. Check for v2 migration before depending. Build `ChartRenderer` interface for swappability. |
| `goingecko` | Replaces archived `go-gecko`. Has built-in rate limiting and retry. |
| `finnhub-go/v2` | REST only. WebSocket needs manual implementation for real-time streaming. |

---

## 7. Error Handling Strategy

### In the TUI (non-blocking)
Errors display in the footer status bar, not as modal popups:
```
Footer: [!] Finnhub: rate limited, using cached data (2m old)
Footer: [!] AI unavailable, using keyword sentiment
```

### In commands (recover from panics)
```go
func safeCmd(fn func() tea.Msg) tea.Cmd {
    return func() tea.Msg {
        defer func() {
            if r := recover(); r != nil {
                // Log to file, return error message
            }
        }()
        return fn()
    }
}
```

### HTTP calls (always with timeout)
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
// Use ctx in all HTTP requests
```

### Logging
- File-based: `~/.local/share/hoard/hoard.log`
- TUI owns stdout — cannot log there
- Structured logging with levels (debug/info/warn/error)
- Debug mode: `hoard --debug` increases verbosity

---

## 8. First-Run Experience

Progressive disclosure, not a setup wizard:

1. **First launch:** Empty portfolio view with helpful message
   ```
   Welcome to Hoard! Press [a] to add your first position.
   Press [?] for help, [Tab] to see your daily cockpit.
   ```

2. **First stock lookup:** "Enter your Finnhub API key (free at finnhub.io):"
   - Saved to config file with restricted permissions (0600)
   - Or via env var: `HOARD_FINNHUB_KEY`

3. **Calendar sidebar:** Shows "Configure calendar: gws auth login or set ics_path in config"

4. **AI features:** Disabled with label "Set HOARD_GROQ_KEY to enable AI summaries"

5. **Config auto-creation:** `~/.config/hoard/config.toml` created on first run with commented defaults

---

## 9. Terminal Compatibility

### Color handling
Bubble Tea v2 auto-detects and downsamples:
- TrueColor (24-bit) → ANSI256 → ANSI (16) → ASCII
- Design palette with ANSI256 as baseline; TrueColor for gradients

### Chart rendering
- Default: braille characters (smooth curves, highest resolution)
- Fallback: block characters (wider terminal support)
- ASCII mode: `--ascii` flag or `theme = "ascii"` in config

### Recommended fonts
Document in README: Cascadia Code, JetBrains Mono, Nerd Fonts
Known issues: Consolas, Courier New, GNOME Terminal "Monospace Regular"

### Testing targets
iTerm2, Terminal.app, Alacritty, kitty, Windows Terminal, VS Code terminal, tmux, SSH sessions

---

## 10. Build & Release

```makefile
VERSION := $(shell git describe --tags --always)

build:
    go build -ldflags "-s -w -X main.version=$(VERSION)" -o bin/hoard ./cmd/hoard

test:
    go test ./... -race -count=1

lint:
    golangci-lint run

release:
    goreleaser release --clean
```

### Distribution
- **GoReleaser** → cross-platform binaries (linux/darwin/windows × amd64/arm64)
- **Homebrew tap** → `brew install yourusername/tap/hoard`
- **AUR** → Arch Linux
- **Go install** → `go install github.com/yourusername/hoard/cmd/hoard@latest`

Single binary, zero runtime dependencies.

---

## 11. Testing Strategy

Layered approach:

| Layer | What | How |
|-------|------|-----|
| **Unit tests** | Business logic (P&L calc, API parsing, DB queries) | Standard Go tests, no Bubble Tea |
| **Model tests** | Each UI component's Update/View | Send messages, assert state changes |
| **Integration tests** | DB operations, API client caching | SQLite in-memory, HTTP mock server |
| **Golden file tests** | Key user flows (add position, switch views) | `teatest` or `catwalk`, run less frequently |
| **Visual regression** | README demo, appearance checks | VHS tape files |

---

## 12. Configuration

```toml
# ~/.config/hoard/config.toml

base_currency = "USD"
theme = "dark"                # "dark" | "light" | "auto" | "ascii"
sidebar_default = "open"      # "open" | "closed"

[market]
stock_provider = "finnhub"
crypto_provider = "coingecko"
refresh_interval_market = "30s"
refresh_interval_crypto = "120s"
refresh_interval_closed = "5m"

[market.finnhub]
api_key = ""  # env: HOARD_FINNHUB_KEY

[calendar]
source = "auto"  # "auto" | "gws" | "google-api" | "ics" | "calcom"
ics_path = ""     # only if source = "ics"

[ai]
provider_priority = ["ollama", "groq", "gemini"]
timeout_ms = 5000
cache_ttl_minutes = 60

[ai.ollama]
endpoint = "http://localhost:11434"
model = "llama3.3"

[ai.groq]
api_key = ""  # env: HOARD_GROQ_KEY

[ai.gemini]
api_key = ""  # env: HOARD_GEMINI_KEY
```

All API keys support both config file and environment variables (env takes precedence).

---

## 13. Known Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| ntcharts not v2-compatible | HIGH | Build ChartRenderer interface. Fallback: asciigraph or custom sparklines. |
| CoinGecko free tier (5-15 req/min) | MEDIUM | Aggressive caching, 120s minimum refresh interval |
| modernc.org/libc version mismatch | MEDIUM | Pin exact version in go.mod. Test after every upgrade. |
| WAL file growth | LOW | `defer rows.Close()` always. Periodic `PRAGMA wal_checkpoint(TRUNCATE)`. |
| Finnhub no WebSocket in Go SDK | MEDIUM | REST polling for MVP. Manual WebSocket implementation later. |
| Command panics break terminal | MEDIUM | Wrap all tea.Cmd functions with recover(). |
| Unofficial APIs (gws, notebooklm-py) | LOW | These are optional integrations with fallbacks. |
