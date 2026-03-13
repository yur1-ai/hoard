# Hoard Implementation Plan — v2 Addendum

> This addendum patches the original plan (`2026-03-13-hoard-implementation-plan.md`).
> It adds new tasks, expands vague ones, inserts two new phases (CI/CD and Distribution),
> adds phase gates, and fixes all issues identified in the independent review.

---

## Updated Phase Structure

| Phase | Name | Tasks | Change |
|-------|------|-------|--------|
| 1 | Foundation | 10 (+4) | Added: logger, CLI flags, safeCmd, v2 API verification |
| 1.5 | **CI/CD Pipeline** | 3 | **NEW PHASE** |
| 2 | Data Layer | 5 (unchanged) | Added phase gate |
| 3 | App Shell & Layout | 8 (+2) | Added: common/table.go, minimum terminal size |
| 4 | Portfolio & Market Data | 9 (+2) | Added: account management, full edit/delete/filter |
| 5 | Watchlist & Charts | 7 (+2) | Added: search overlay, candlestick charts |
| 6 | News & Sidebar | 7 (+1) | Added: CryptoPanic news source |
| 6.5 | **Distribution & Packaging** | 4 | **NEW PHASE** — MVP release point |
| 7 | AI Layer | 6 (+1) | Added: token budget tracking |
| 8 | Polish & Release | 6 (+1) | Added: golden file tests |

**Total: 65 tasks** (was 45)

---

## Phase Gate Protocol

**EVERY phase ends with a mandatory gate. Do NOT proceed until the gate passes.**

```
Phase Gate Checklist:
  □ All tests pass: `go test ./... -race -count=1`
  □ Lint passes: `golangci-lint run` (after Phase 1.5 sets it up)
  □ Build succeeds: `make build`
  □ Manual smoke test: launch app, exercise new features, quit cleanly
  □ No regressions: features from prior phases still work
  □ CI green (after Phase 1.5): push branch, verify GitHub Actions pass
  □ Commit all work with descriptive message
```

Include this checklist as the LAST task of every phase. An engineer must check every box before starting the next phase.

---

## BLOCKING FIX 1: Bubble Tea v2 API Verification

### Task 1.0: Verify Bubble Tea v2 API Before Writing Code

**This must be the VERY FIRST task. All subsequent code depends on it.**

- [ ] **Step 1: Check actual v2 release status**

```bash
# Check if v2 is published
go list -m -versions charm.land/bubbletea/v2 2>/dev/null || echo "v2 not yet on vanity domain"
# If not on vanity domain, check GitHub
curl -s https://api.github.com/repos/charmbracelet/bubbletea/releases | grep tag_name | head -5
```

- [ ] **Step 2: Read the actual v2 upgrade guide**

Fetch and read: https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md

Verify these critical assumptions from our plan:
1. Does `View()` return `tea.View` struct or `string`?
2. What is the actual `KeyPressMsg` field for key identity? (`Code`? `Key`? `String()`?)
3. What is the import path? (`charm.land/bubbletea/v2` or `github.com/charmbracelet/bubbletea/v2`?)
4. Does `tea.Tick` still work the same way?

- [ ] **Step 3: Create a minimal test program**

```go
// /tmp/bt2test/main.go — throwaway test
package main

import (
    "fmt"
    tea "charm.land/bubbletea/v2" // or correct import
)

type model struct{}

func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        fmt.Printf("Key: %+v\n", msg) // see actual field names
        return m, tea.Quit
    }
    return m, nil
}
func (m model) View() ??? { // fill in actual return type
    return ???
}

func main() {
    tea.NewProgram(model{}).Run()
}
```

Build and run. Press a key. Note exact field names and types.

- [ ] **Step 4: Document actual API in a reference file**

Create `docs/bubble-tea-v2-api-reference.md` with verified field names, types, and patterns.
Update ALL code snippets in the plan if they differ from the verified API.

- [ ] **Step 5: If v2 is NOT yet released as a stable version**

Options:
1. Use a specific commit/pre-release tag: `go get charm.land/bubbletea/v2@v2.0.0-beta.1`
2. Fall back to v1 and plan a migration later (changes import paths + View signature)

**Decision must be documented before proceeding.**

---

## BLOCKING FIX 2: File-Based Logger

### Task 1.7: Set Up Structured File Logger

**Files:**
- Create: `internal/logger/logger.go`

- [ ] **Step 1: Implement logger using stdlib `log/slog`**

```go
// internal/logger/logger.go
package logger

import (
    "io"
    "log/slog"
    "os"
    "path/filepath"

    "github.com/yourusername/hoard/internal/config"
)

var Log *slog.Logger

// Init sets up file-based logging. Must be called before any log usage.
func Init(debug bool) error {
    logPath := config.LogFilePath()
    if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
        return err
    }

    f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return err
    }

    level := slog.LevelInfo
    if debug {
        level = slog.LevelDebug
    }

    handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: level})
    Log = slog.New(handler)
    slog.SetDefault(Log)
    return nil
}

// Discard sets up a no-op logger (for tests).
func Discard() {
    Log = slog.New(slog.NewTextHandler(io.Discard, nil))
    slog.SetDefault(Log)
}
```

- [ ] **Step 2: Replace ALL `log.Printf` in store/db.go with `slog.Info`**

```go
// BEFORE (corrupts TUI):
log.Printf("migration applied: %s", entry.Name())

// AFTER (file-only):
slog.Info("migration applied", "file", entry.Name())
```

- [ ] **Step 3: Call `logger.Init()` in main.go BEFORE store.Open()**

- [ ] **Step 4: Commit**

```bash
git add internal/logger/ internal/store/db.go cmd/hoard/main.go
git commit -m "feat: file-based structured logging with slog, replace stdout logging"
```

---

## BLOCKING FIX 3: CLI Flag Parsing

### Task 1.8: CLI Flags

**Files:**
- Modify: `cmd/hoard/main.go`

- [ ] **Step 1: Add flag parsing with stdlib `flag` package**

```go
// cmd/hoard/main.go — add before config loading
var (
    flagVersion bool
    flagDebug   bool
    flagASCII   bool
    flagConfig  string
)

func init() {
    flag.BoolVar(&flagVersion, "version", false, "Print version and exit")
    flag.BoolVar(&flagDebug, "debug", false, "Enable debug logging")
    flag.BoolVar(&flagASCII, "ascii", false, "Use ASCII-only rendering")
    flag.StringVar(&flagConfig, "config", "", "Path to config file")
}

func main() {
    flag.Parse()

    if flagVersion {
        fmt.Printf("hoard %s\n", version)
        os.Exit(0)
    }

    // Init logger first (needs debug flag)
    if err := logger.Init(flagDebug); err != nil {
        fmt.Fprintf(os.Stderr, "logger error: %v\n", err)
        os.Exit(1)
    }

    // Load config (respect --config flag)
    var cfg config.Config
    var err error
    if flagConfig != "" {
        cfg, err = config.LoadFromFile(flagConfig)
    } else {
        cfg, err = config.Load()
    }
    // ...

    if flagASCII {
        cfg.Theme = "ascii"
    }
    // ... rest of main
}
```

- [ ] **Step 2: Commit**

```bash
git add cmd/hoard/main.go
git commit -m "feat: CLI flags for --version, --debug, --ascii, --config"
```

---

## BLOCKING FIX 4: safeCmd Panic Recovery

### Task 1.9: safeCmd Wrapper

**Files:**
- Create: `internal/app/safecmd.go`
- Create: `internal/app/safecmd_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestSafeCmdRecoversPanic(t *testing.T) {
    panicCmd := safeCmd(func() tea.Msg {
        panic("test panic")
    })
    msg := panicCmd() // should NOT panic
    errMsg, ok := msg.(ErrMsg)
    if !ok {
        t.Fatal("expected ErrMsg from recovered panic")
    }
    if errMsg.Context != "panic" {
        t.Errorf("expected panic context, got %s", errMsg.Context)
    }
}

func TestSafeCmdPassesThrough(t *testing.T) {
    normalCmd := safeCmd(func() tea.Msg {
        return TickMsg(time.Now())
    })
    msg := normalCmd()
    if _, ok := msg.(TickMsg); !ok {
        t.Fatal("expected TickMsg passthrough")
    }
}
```

- [ ] **Step 2: Run test — verify it fails**
- [ ] **Step 3: Implement safeCmd**

```go
// internal/app/safecmd.go
package app

import (
    "fmt"
    "log/slog"

    tea "charm.land/bubbletea/v2"
)

// safeCmd wraps a tea.Cmd function with panic recovery.
// If the function panics, it returns an ErrMsg instead of crashing the terminal.
func safeCmd(fn func() tea.Msg) tea.Cmd {
    return func() tea.Msg {
        defer func() {
            if r := recover(); r != nil {
                slog.Error("panic in command", "recover", r)
            }
        }()
        return fn()
    }
}
```

Wait — the recover needs to actually return. Use a named return:

```go
func safeCmd(fn func() tea.Msg) tea.Cmd {
    return func() (result tea.Msg) {
        defer func() {
            if r := recover(); r != nil {
                slog.Error("panic in command", "recover", r)
                result = ErrMsg{
                    Err:     fmt.Errorf("internal error: %v", r),
                    Context: "panic",
                }
            }
        }()
        return fn()
    }
}
```

- [ ] **Step 4: Run test — verify it passes**
- [ ] **Step 5: Commit**

```bash
git add internal/app/safecmd.go internal/app/safecmd_test.go
git commit -m "feat: safeCmd panic recovery wrapper for all tea.Cmd functions"
```

**RULE: Every `tea.Cmd` created in the codebase MUST use `safeCmd`. Search for raw `func() tea.Msg` during code review.**

---

## NEW PHASE 1.5: CI/CD Pipeline

**Goal:** GitHub Actions for lint + test on every push/PR. Set up once, benefits every subsequent phase.

### Task 1.5.1: golangci-lint Configuration

**Files:**
- Create: `.golangci.yml`

- [ ] **Step 1: Write golangci-lint config**

```yaml
# .golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gosimple
    - gocritic
    - misspell
    - revive
    - bodyclose
    - contextcheck
  disable:
    - depguard

linters-settings:
  revive:
    rules:
      - name: unexported-return
        disabled: true

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

- [ ] **Step 2: Run lint locally**

```bash
golangci-lint run
# Fix any issues found
```

- [ ] **Step 3: Commit**

```bash
git add .golangci.yml
git commit -m "chore: golangci-lint configuration"
```

---

### Task 1.5.2: GitHub Actions Workflow

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Write CI workflow**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Build
        run: go build -v ./...

      - name: Test
        run: go test -race -count=1 -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

- [ ] **Step 2: Create GitHub repo and push**

```bash
# Create repo on GitHub first (via gh CLI or web UI)
gh repo create hoard --public --source=. --remote=origin
git push -u origin main
```

- [ ] **Step 3: Verify CI runs and passes**

```bash
# Wait ~60s then check
gh run list --limit 3
# Expected: workflow running or passed
```

- [ ] **Step 4: Commit**

```bash
git add .github/
git commit -m "ci: GitHub Actions workflow for test + lint"
```

---

### Task 1.5.3: Add CI Badge to README Stub

**Files:**
- Create: `README.md` (minimal stub — full README in Phase 8)

- [ ] **Step 1: Write minimal README**

```markdown
# Hoard

Terminal finance dashboard & daily cockpit.

[![CI](https://github.com/yourusername/hoard/actions/workflows/ci.yml/badge.svg)](https://github.com/yourusername/hoard/actions/workflows/ci.yml)

> Work in progress. See `docs/` for design and architecture.
```

- [ ] **Step 2: Commit and push**

```bash
git add README.md
git commit -m "docs: minimal README with CI badge"
git push
```

---

## EXPANDED: Task 3.7 (NEW): Reusable Table Component

**Files:**
- Create: `internal/ui/common/table.go`

- [ ] **Step 1: Create styled table wrapper**

Wraps `bubbles/v2/table` with project-specific defaults: Lip Gloss borders, green/red coloring for financial values, column alignment (right-align numbers), and vim keybindings (j/k).

```go
// internal/ui/common/table.go
package common

import "charm.land/bubbles/v2/table"

// NewStyledTable returns a table.Model with Hoard's default styling.
func NewStyledTable(columns []table.Column, rows []table.Row, height int) table.Model {
    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithHeight(height),
        table.WithFocused(true),
    )
    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(ColorMuted).
        BorderBottom(true).
        Bold(true)
    s.Selected = s.Selected.
        Foreground(ColorWhite).
        Background(ColorHighlight)
    t.SetStyles(s)
    return t
}

// FormatMoney returns a right-aligned, colored money string.
func FormatMoney(value float64) string { ... }

// FormatPercent returns a colored percentage string (green positive, red negative).
func FormatPercent(value float64) string { ... }

// FormatChange returns "▲ +2.3%" or "▼ -1.2%" with appropriate color.
func FormatChange(value float64) string { ... }
```

- [ ] **Step 2: Commit**

```bash
git add internal/ui/common/table.go
git commit -m "feat(ui): reusable styled table with financial formatting helpers"
```

---

## EXPANDED: Task 3.8 (NEW): Minimum Terminal Size Check

**Files:**
- Modify: `internal/app/update.go`

- [ ] **Step 1: Add minimum size constants and check**

```go
const (
    minWidth  = 80
    minHeight = 24
)

// In Update(), handle WindowSizeMsg:
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.tooSmall = msg.Width < minWidth || msg.Height < minHeight
    // propagate to all sub-models...

// In View():
if m.tooSmall {
    return tea.View{Content: lipgloss.Place(
        m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        fmt.Sprintf("Terminal too small (%dx%d).\nMinimum: %dx%d.\nPlease resize.",
            m.width, m.height, minWidth, minHeight),
    )}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/app/
git commit -m "feat: minimum terminal size check with friendly resize message"
```

---

## EXPANDED: Task 4.0 (NEW): Account Management

**Files:**
- Modify: `internal/ui/market/portfolio/model.go`
- Modify: `internal/store/portfolio.go`

- [ ] **Step 1: Auto-create default account on first run**

In `main.go`, after DB init:
```go
accounts, _ := store.ListAccounts(db)
if len(accounts) == 0 {
    store.CreateAccount(db, "Default", "brokerage", cfg.BaseCurrency)
    slog.Info("created default account")
}
```

- [ ] **Step 2: Add account selector to add-position form**

If multiple accounts exist, show a selector. If only one, auto-select it.

- [ ] **Step 3: Add `[A]` key for "Manage Accounts" dialog**

Simple form: create account (name, type, currency) or delete empty account.

- [ ] **Step 4: Commit**

```bash
git add internal/ cmd/
git commit -m "feat: account management with default account auto-creation"
```

---

## EXPANDED: Task 4.6: Full Portfolio CRUD (Edit, Delete, Filter)

Patch the existing Task 4.6 to add these MISSING keybinds:

- [ ] **Step: Add `[e]dit` position** — opens pre-filled form for selected holding. Update quantity and avg cost. Save via store.UpdateHolding.

- [ ] **Step: Add `[d]elete` position** — confirmation prompt ("Delete AAPL? y/n"), then store.DeleteHolding.

- [ ] **Step: Add `[f]ilter` by account** — cycle through accounts (All → Account1 → Account2 → All). Filter the holdings table.

- [ ] **Step: Add `[r]efresh`** — force immediate market data fetch (bypass cache TTL).

- [ ] **Step: Wire inline sparklines** — use ChartRenderer.RenderSparkline for each row in the holdings table. Feed intraday price data from cache.

- [ ] **Step: Commit**

```bash
git commit -m "feat(ui): portfolio edit, delete, filter by account, manual refresh, inline sparklines"
```

---

## EXPANDED: Task 4.7: Tick Loop (was vague — now concrete)

Replace the vague "Follow the pattern" with explicit steps:

- [ ] **Step 1: Implement `isUSMarketOpen()`**

```go
// internal/app/market_hours.go
package app

import "time"

var easternTZ *time.Location

func init() {
    var err error
    easternTZ, err = time.LoadLocation("America/New_York")
    if err != nil {
        easternTZ = time.FixedZone("EST", -5*60*60)
    }
}

func isUSMarketOpen() bool {
    now := time.Now().In(easternTZ)
    weekday := now.Weekday()
    if weekday == time.Saturday || weekday == time.Sunday {
        return false
    }
    hour, min := now.Hour(), now.Minute()
    minuteOfDay := hour*60 + min
    // Market open: 9:30 (570) to 16:00 (960)
    return minuteOfDay >= 570 && minuteOfDay < 960
}

func isExtendedHours() bool {
    now := time.Now().In(easternTZ)
    weekday := now.Weekday()
    if weekday == time.Saturday || weekday == time.Sunday {
        return false
    }
    hour, min := now.Hour(), now.Minute()
    minuteOfDay := hour*60 + min
    // Pre-market: 4:00 (240) to 9:30 (570)
    // Post-market: 16:00 (960) to 20:00 (1200)
    return (minuteOfDay >= 240 && minuteOfDay < 570) ||
           (minuteOfDay >= 960 && minuteOfDay < 1200)
}
```

- [ ] **Step 2: Write test for market hours**

```go
func TestMarketOpenDuringTrading(t *testing.T)    // Tuesday 10:00 ET → true
func TestMarketClosedWeekend(t *testing.T)        // Saturday → false
func TestExtendedHoursPreMarket(t *testing.T)     // Tuesday 7:00 ET → true
```

- [ ] **Step 3: Implement NeedsRefresh tracker**

```go
type refreshTracker struct {
    lastRefresh map[string]time.Time // "equity", "crypto", "news", "calendar", "currency"
}

func (r *refreshTracker) NeedsRefresh(key string, maxAge time.Duration) bool {
    last, ok := r.lastRefresh[key]
    return !ok || time.Since(last) > maxAge
}

func (r *refreshTracker) MarkRefreshed(key string) {
    r.lastRefresh[key] = time.Now()
}
```

- [ ] **Step 4: Wire into App.Update TickMsg handler**

```go
case TickMsg:
    cmds := []tea.Cmd{m.scheduleTick()}

    // Equity refresh
    equityInterval := 5 * time.Minute // closed
    if isUSMarketOpen() {
        equityInterval = 30 * time.Second
    } else if isExtendedHours() {
        equityInterval = 2 * time.Minute
    }
    if m.refresh.NeedsRefresh("equity", equityInterval) {
        symbols := m.store.AllEquitySymbols()
        cmds = append(cmds, safeCmd(func() tea.Msg {
            // fetch with timeout
        }))
    }

    // Crypto (24/7, every 120s)
    if m.refresh.NeedsRefresh("crypto", 120*time.Second) { ... }

    // News (every 5 min)
    if m.refresh.NeedsRefresh("news", 5*time.Minute) { ... }

    // Calendar (every 15 min, only if sidebar visible)
    if m.sidebarOpen && m.refresh.NeedsRefresh("calendar", 15*time.Minute) { ... }

    // Currency (daily)
    if m.refresh.NeedsRefresh("currency", 24*time.Hour) { ... }

    // WAL checkpoint (every 30 min)
    if m.refresh.NeedsRefresh("wal_checkpoint", 30*time.Minute) {
        cmds = append(cmds, safeCmd(func() tea.Msg {
            store.WALCheckpoint(m.db)
            return nil
        }))
    }

    return m, tea.Batch(cmds...)
```

- [ ] **Step 5: Commit**

```bash
git add internal/app/
git commit -m "feat: adaptive tick loop with market hours, timezone, refresh tracking, WAL checkpoint"
```

---

## EXPANDED: Task 5.6 (NEW): Search Overlay

**Files:**
- Create: `internal/ui/search/model.go`
- Modify: `internal/app/update.go`

- [ ] **Step 1: Implement search overlay**

When user presses `/` in modeNormal:
1. Switch to `modeSearch`
2. Show text input overlay at top of screen
3. Fuzzy-match against: portfolio symbols, watchlist symbols, all known symbols
4. Results update as user types
5. Enter selects result → navigate to that symbol (portfolio or charts view)
6. Escape closes overlay → back to modeNormal

Use `bubbles/v2/textinput` for the search field. Simple substring matching for MVP (fuzzy matching enhancement later).

- [ ] **Step 2: Commit**

```bash
git add internal/ui/search/ internal/app/
git commit -m "feat(ui): search overlay with symbol matching"
```

---

## EXPANDED: Task 5.7 (NEW): Candlestick Chart Mode

- [ ] **Step 1: Add candlestick rendering to ChartRenderer interface**

```go
type ChartRenderer interface {
    RenderLineChart(data ChartData, width, height int) string
    RenderCandlestick(candles []Candle, width, height int) string
    RenderSparkline(values []float64, width int) string
}
```

- [ ] **Step 2: Implement ASCII candlestick renderer**

Use half-block characters (▄▀) for candle bodies, pipe (│) for wicks. Green for up candles, red for down.

- [ ] **Step 3: Add chart type toggle in charts view**

`[c]` key cycles: Line → Candlestick → Line.

- [ ] **Step 4: Commit**

```bash
git commit -m "feat(ui): candlestick chart rendering with toggle"
```

---

## EXPANDED: Task 6.1: Add CryptoPanic News Source

Patch Task 6.1 to include crypto-specific news:

- [ ] **Step: Create `internal/service/market/cryptopanic.go`**

CryptoPanic API: `https://cryptopanic.com/api/v1/posts/?auth_token=FREE&currencies=BTC,ETH`
Free tier: no auth token needed for public posts, 5 req/min.

Implement as a secondary news source. When fetching news for crypto symbols, query CryptoPanic. For equity symbols, query Finnhub news.

- [ ] **Step: Commit**

```bash
git commit -m "feat(service): CryptoPanic news source for crypto holdings"
```

---

## EXPANDED: Task 6.3: Fix gws Command

Replace the inconsistent command. Architecture doc says `gws calendar +agenda`, plan said `gws calendar events list --format json`. The correct approach:

```go
// Try the recipe command first (human-readable, structured)
cmd := exec.CommandContext(ctx, "gws", "calendar", "+agenda", "--format", "json")
output, err := cmd.Output()
if err != nil {
    // Fallback to raw API command
    cmd = exec.CommandContext(ctx, "gws", "calendar", "events", "list",
        "--calendarId", "primary",
        "--timeMin", from.Format(time.RFC3339),
        "--timeMax", to.Format(time.RFC3339),
    )
    output, err = cmd.Output()
}
```

Also add execution timeout:
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
```

---

## NEW PHASE 6.5: Distribution & Packaging

**Goal:** Publish the MVP. Users can `brew install` and start using Hoard.

> **Placement rationale:** After Phase 6, all core features work (portfolio, watchlist, charts, news, sidebar). This is the MVP. AI (Phase 7) and polish (Phase 8) are enhancements. Publishing here gets early users and feedback.

### Package Manager Research Summary

| Manager | Platform | GoReleaser Support | Effort | Audience |
|---------|----------|-------------------|--------|----------|
| **Homebrew** | macOS + Linux | Built-in (brews section) | Low | Primary target — most Go CLI users on macOS |
| **Scoop** | Windows | Built-in (scoops section) | Low | Windows developers |
| **AUR** | Arch Linux | Built-in (aurs section) | Low | Arch/Manjaro users |
| **Nix/NUR** | NixOS + any | Built-in (nix section) | Medium | Nix enthusiasts |
| **go install** | Any with Go | Free (just works) | Zero | Go developers |
| **Binary download** | Any | Built-in (GitHub Releases) | Zero | Everyone |
| **Snap** | Ubuntu | Via nfpm plugin | Medium | Ubuntu users |
| **Deb/RPM** | Debian/RHEL | Via nfpm plugin | Medium | Server admins |

**Recommendation:** Start with Homebrew + Scoop + AUR + `go install` + binary download. All are built into GoReleaser with minimal config. Skip Snap/Deb/RPM/Nix for now — add later if requested.

### Task 6.5.1: Create Homebrew Tap Repository

- [ ] **Step 1: Create the tap repo on GitHub**

```bash
gh repo create homebrew-tap --public --description "Homebrew formulae for personal tools"
```

This creates `github.com/yourusername/homebrew-tap`. GoReleaser will automatically push formulae here.

- [ ] **Step 2: Create a GitHub Personal Access Token**

Go to https://github.com/settings/tokens → Generate new token (classic):
- Scopes: `repo` (full control)
- Name: "goreleaser-homebrew"
- Save as `GITHUB_TOKEN` secret in the hoard repo

```bash
gh secret set HOMEBREW_TAP_TOKEN < /dev/stdin
# Paste token, Ctrl+D
```

- [ ] **Step 3: Commit (nothing to commit in hoard repo yet — tap repo is separate)**

---

### Task 6.5.2: GoReleaser Configuration (Full)

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Write full GoReleaser config**

```yaml
# .goreleaser.yaml
version: 2

before:
  hooks:
    - go mod tidy
    - go test ./...

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
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}

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
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    homepage: "https://github.com/yourusername/hoard"
    description: "Terminal finance dashboard & daily cockpit"
    license: "MIT"
    install: |
      bin.install "hoard"
    test: |
      system "#{bin}/hoard", "--version"
    caveats: |
      To get started:
        hoard

      For stock data, get a free API key at https://finnhub.io
      then set: export HOARD_FINNHUB_KEY=your_key

scoops:
  - repository:
      owner: yourusername
      name: scoop-bucket
    homepage: "https://github.com/yourusername/hoard"
    description: "Terminal finance dashboard & daily cockpit"
    license: MIT

aurs:
  - name: hoard-bin
    homepage: "https://github.com/yourusername/hoard"
    description: "Terminal finance dashboard & daily cockpit"
    maintainers:
      - "Your Name <you@example.com>"
    license: "MIT"
    private_key: "{{ .Env.AUR_KEY }}"
    git_url: "ssh://aur@aur.archlinux.org/hoard-bin.git"

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^chore:"
      - "^test:"
      - "^ci:"

release:
  github:
    owner: yourusername
    name: hoard
  draft: false
  prerelease: auto
```

- [ ] **Step 2: Validate config**

```bash
goreleaser check
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yaml
git commit -m "chore: GoReleaser config for Homebrew, Scoop, AUR, and binary releases"
```

---

### Task 6.5.3: Release GitHub Action

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Write release workflow (triggered by tag push)**

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: release workflow with GoReleaser on tag push"
```

---

### Task 6.5.4: First Release (MVP)

- [ ] **Step 1: Create and push tag**

```bash
git tag -a v0.1.0 -m "v0.1.0: MVP — portfolio, watchlist, charts, news, daily cockpit"
git push origin v0.1.0
```

- [ ] **Step 2: Verify release action runs**

```bash
# Wait ~60s
gh run list --limit 3
# Wait for completion
gh run view --log  # check for errors
```

- [ ] **Step 3: Verify Homebrew formula was pushed to tap repo**

```bash
gh api repos/yourusername/homebrew-tap/contents/Formula/hoard.rb --jq '.name'
# Expected: "hoard.rb"
```

- [ ] **Step 4: Test installation**

```bash
brew tap yourusername/tap
brew install hoard
hoard --version
# Expected: "hoard v0.1.0"
```

- [ ] **Step 5: Verify binary downloads on GitHub Releases page**

Check: `https://github.com/yourusername/hoard/releases/tag/v0.1.0`
Should have: darwin_amd64, darwin_arm64, linux_amd64, linux_arm64, windows_amd64, checksums.txt

---

## EXPANDED: Task 7.6 (NEW): AI Token Budget Tracking

**Files:**
- Modify: `internal/service/ai/chain.go`
- Modify: `internal/ui/header/model.go`

- [ ] **Step 1: Add quota tracking to chain**

```go
type quotaTracker struct {
    provider   string
    used       int
    limit      int  // 0 = unknown/unlimited
    resetAt    time.Time
}

func (c *Chain) Quota() quotaTracker { return c.quota }
```

For Groq: parse `x-ratelimit-remaining` response header.
For Gemini: parse rate limit headers if available.
For Ollama: unlimited, skip tracking.

- [ ] **Step 2: Display in header**

```
AI: Groq ● 847/1000
```

When quota < 10%: turn indicator yellow. When exhausted: auto-switch provider, show:
```
AI: Gemini ● (Groq quota exhausted)
```

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(ai): token budget tracking with header display and auto-switch"
```

---

## EXPANDED: Task 8.6 (NEW): Golden File Tests

**Files:**
- Create: `internal/app/golden_test.go`

- [ ] **Step 1: Install teatest**

```bash
go get charm.land/x/exp/teatest
```

- [ ] **Step 2: Write golden file tests for key flows**

```go
func TestGolden_EmptyPortfolio(t *testing.T)     // first-run state
func TestGolden_ViewSwitching(t *testing.T)       // 1→2→3→4 view labels
func TestGolden_SidebarToggle(t *testing.T)       // Tab opens/closes sidebar
```

Run with `-update` flag to generate golden files:
```bash
go test ./internal/app/ -run TestGolden -update
```

Then run without `-update` to verify against golden files:
```bash
go test ./internal/app/ -run TestGolden
```

- [ ] **Step 3: Add golden files to git**

```bash
git add internal/app/testdata/
git commit -m "test: golden file tests for core UI flows"
```

---

## EXPANDED: Config Validation and Permissions

Add these steps to Phase 1, Task 1.3:

- [ ] **Step: Add Validate() method**

```go
func (c Config) Validate() error {
    validThemes := map[string]bool{"dark": true, "light": true, "auto": true, "ascii": true}
    if !validThemes[c.Theme] {
        return fmt.Errorf("invalid theme %q: must be dark, light, auto, or ascii", c.Theme)
    }
    validSidebar := map[string]bool{"open": true, "closed": true}
    if !validSidebar[c.SidebarDefault] {
        return fmt.Errorf("invalid sidebar_default %q: must be open or closed", c.SidebarDefault)
    }
    // Validate durations parse correctly
    if _, err := time.ParseDuration(c.Market.RefreshIntervalMarket); err != nil {
        return fmt.Errorf("invalid refresh_interval_market: %w", err)
    }
    return nil
}
```

- [ ] **Step: Write config with 0600 permissions**

```go
func WriteConfigFile(path string, data []byte) error {
    return os.WriteFile(path, data, 0600) // restrictive: API keys inside
}
```

- [ ] **Step: Add test for config validation**

```go
func TestValidateRejectsInvalid(t *testing.T) {
    cfg := DefaultConfig()
    cfg.Theme = "rainbow"
    if err := cfg.Validate(); err == nil {
        t.Error("expected validation error for invalid theme")
    }
}
```

---

## No-Internet + No-Cache First Launch

Add to Task 8.1:

- [ ] **Step: Handle zero-data state gracefully**

When portfolio has holdings but cache has NO prices AND network is down:
```
 AAPL    50 shares    Avg: $165.20    Price: --    Day: --    P&L: --
                                       ⚠ No market data (offline)
```

Show `--` for unknown values instead of `$0.00`. Show a non-blocking footer warning: `[!] No market data available. Check internet connection.`

---

## Summary of All Changes

### New Tasks Added (20)
| Task | Phase | Description |
|------|-------|-------------|
| 1.0 | 1 | Bubble Tea v2 API verification |
| 1.7 | 1 | File-based logger (slog) |
| 1.8 | 1 | CLI flag parsing |
| 1.9 | 1 | safeCmd panic recovery |
| 1.5.1 | 1.5 | golangci-lint config |
| 1.5.2 | 1.5 | GitHub Actions CI workflow |
| 1.5.3 | 1.5 | README stub with CI badge |
| 3.7 | 3 | Reusable table component |
| 3.8 | 3 | Minimum terminal size check |
| 4.0 | 4 | Account management + default account |
| 5.6 | 5 | Search overlay |
| 5.7 | 5 | Candlestick chart mode |
| 6.5.1 | 6.5 | Homebrew tap repository |
| 6.5.2 | 6.5 | GoReleaser full config |
| 6.5.3 | 6.5 | Release GitHub Action |
| 6.5.4 | 6.5 | First MVP release |
| 7.6 | 7 | AI token budget tracking |
| 8.6 | 8 | Golden file tests |
| Every phase | All | Phase gate checklist |
| Every phase | All | Unit test validation before proceeding |

### Existing Tasks Expanded (7)
| Task | What Changed |
|------|-------------|
| 4.6 | Added edit, delete, filter, refresh, sparklines |
| 4.7 | Full tick loop code: market hours, timezone, NeedsRefresh, WAL checkpoint |
| 6.1 | Added CryptoPanic news source for crypto |
| 6.3 | Fixed gws command, added timeout |
| 1.3 | Added config validation + 0600 permissions |
| 8.1 | Added no-internet + no-cache handling |
| store/db.go | Replaced log.Printf with slog |

### Bugs Fixed (3)
| Bug | Fix |
|-----|-----|
| log.Printf corrupts TUI | Replaced with file-based slog |
| safeCmd never implemented | New Task 1.9 |
| WALCheckpoint never called | Integrated into tick loop |
