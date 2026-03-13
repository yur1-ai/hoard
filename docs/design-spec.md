# Hoard — Terminal Finance Dashboard & Daily Cockpit

**Date:** 2026-03-13
**Status:** Design Complete — Pending Implementation Plan
**Stack:** Go + Bubble Tea v2 + SQLite + Optional AI

---

## 1. Overview

Hoard is a terminal-based (TUI) personal finance dashboard with a lightweight developer daily cockpit. It combines real-time stock/ETF/crypto portfolio tracking with a collapsible sidebar featuring calendar agenda, task list, and standup logger.

**Core principles:**
- Local-first: single SQLite file, no cloud account required
- AI-optional: every feature works without AI; AI enhances but never gates functionality
- Terminal-native: Vim keybindings, works over SSH, single binary distribution
- Graceful degradation: API failures fall back to cached data; AI failures fall back to non-AI alternatives

**Target user:** Active investor (long-term + swing trading) who lives in the terminal and wants to reduce context-switching between finance apps, calendar, and task tools.

---

## 2. Layout Architecture

### Primary: Market dashboard with collapsible sidebar

```
┌─────────────────────────────────────────────┬──────────────────┐
│  Portfolio / Watchlist / Charts / News       │  Daily           │
│  ─────────────────────────────────────────── │  ──────────────  │
│                                              │  TODAY Mar 13    │
│  AAPL    $198.42  ▲ +2.3%  ██████████▎      │  09:00 Standup   │
│  GOOGL   $171.20  ▼ -0.8%  ████████▌        │  11:00 1:1 w/Pat │
│  MSFT    $428.15  ▲ +1.1%  █████████▍       │  14:00 Sprint    │
│  TSLA    $245.80  ▼ -3.2%  ██████▊          │  ──────────────  │
│  NVDA    $892.60  ▲ +4.5%  ████████████▌    │  TASKS           │
│                                              │  ☐ Fix auth bug  │
│  ───────────────────────── P&L ───────────── │  ☑ Review PR     │
│  Today:  +$342.18 (+0.8%)                    │  ☐ Deploy v2.1   │
│  Total:  +$4,218.50 (+12.3%)                │  ──────────────  │
│                                              │  STANDUP         │
│  ─────────────────────── News ────────────── │  Y: Shipped API  │
│  "Fed signals rate pause..." — 10m ago       │  T: Fix auth     │
│  "NVDA earnings beat..." — 2h ago            │  B: Waiting on   │
│                                              │     staging env  │
└─────────────────────────────────────────────┴──────────────────┘
 [1]Portfolio [2]Watchlist [3]Charts [4]News    [Tab]Toggle sidebar
```

**Sidebar collapsed:** Market dashboard goes full-width.
**Toggle:** `Tab` key. Default state configurable (`sidebar_default = "open" | "closed"`).
**Focus:** `Tab` cycles focus between market area and sidebar. Arrow keys navigate within focused panel.
**Sidebar width:** ~25-30 columns fixed.

---

## 3. Data Model

**Storage:** Single SQLite file at `~/.local/share/hoard/hoard.db`
**Config:** `~/.config/hoard/config.toml` (XDG-compliant)

```sql
-- Accounts
CREATE TABLE accounts (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('brokerage','retirement','crypto_wallet')),
    currency TEXT NOT NULL DEFAULT 'USD',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Holdings (current positions, derived from transactions)
CREATE TABLE holdings (
    id INTEGER PRIMARY KEY,
    account_id INTEGER REFERENCES accounts(id),
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    quantity REAL NOT NULL,
    avg_cost_basis REAL NOT NULL,
    notes TEXT
);

-- Transactions
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY,
    account_id INTEGER REFERENCES accounts(id),
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    type TEXT NOT NULL CHECK(type IN ('buy','sell','dividend','transfer')),
    quantity REAL NOT NULL,
    price REAL NOT NULL,
    fee REAL DEFAULT 0,
    date DATETIME NOT NULL,
    notes TEXT
);

-- Watchlists
CREATE TABLE watchlists (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE watchlist_items (
    watchlist_id INTEGER REFERENCES watchlists(id),
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (watchlist_id, symbol)
);

-- Market cache (avoid API rate limits)
CREATE TABLE market_cache (
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

-- Currency rates (refreshed daily)
CREATE TABLE currency_rates (
    from_currency TEXT NOT NULL,
    to_currency TEXT NOT NULL,
    rate REAL NOT NULL,
    fetched_at DATETIME NOT NULL,
    PRIMARY KEY (from_currency, to_currency)
);

-- Tasks (simple todo)
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'todo' CHECK(status IN ('todo','doing','done')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- Standup entries
CREATE TABLE standup_entries (
    id INTEGER PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    yesterday TEXT,
    today TEXT,
    blockers TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

-- Calendar cache
CREATE TABLE calendar_cache (
    event_id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    calendar_name TEXT,
    last_synced DATETIME
);
```

### Asset type handling
- Stocks/ETFs and crypto share the same tables
- Distinguished by `market` field (`us_equity` | `crypto`)
- Crypto symbols stored with suffix convention (e.g., `BTC-USD`)

### Portfolio entry
- MVP: Manual entry (add position: symbol, quantity, cost basis)
- Fast-follow: CSV import from Robinhood and other brokerages
- Future: Robinhood Crypto API for real-time crypto position sync

### Currency conversion
- `base_currency` in config (display everything in this currency)
- Frankfurter API: free, no key, ECB data, refreshed once daily
- Individual holdings show native currency with converted value if different

---

## 4. Market Dashboard Features

### View 1: Portfolio [1] (default)

| Column | Description |
|--------|-------------|
| Symbol | Ticker symbol |
| Shares | Quantity held |
| Avg Cost | Average cost basis per share |
| Price | Current market price |
| Day Chg | Intraday change % with color |
| P&L | Unrealized profit/loss |
| Alloc | Portfolio allocation % |

- Filter by account or view all combined
- Sort by any column
- Inline sparklines for intraday movement
- Color-coded: green gains, red losses
- Keys: `[a]dd` `[e]dit` `[d]elete` `[f]ilter` `[/]search` `[r]efresh`

### View 2: Watchlist [2]

- Multiple named watchlists (e.g., "Swing Candidates", "Long-Term Research")
- Quick-add symbol with fuzzy search
- "Move to portfolio" shortcut when entering a trade
- Keys: `[n]ew watchlist` `[a]dd symbol` `[d]elete` `[m]ove to portfolio`

### View 3: Charts [3]

- Line chart with price history (braille characters for smooth curves)
- Timeframe selection: 1D, 1W, 1M, 3M, 1Y, ALL
- Basic technical indicators: MA(50), MA(200), RSI(14), Volume bars
- Works for both stocks and crypto
- Candlestick/OHLC rendering when terminal supports it

### View 4: News [4]

- Auto-filtered to symbols in portfolio + watchlists
- Sentiment indicator (green/red/neutral dot)
- `[s]` triggers AI summarization (optional)
- `[o]` opens full article in default browser
- `[w]` on a holding: "Why did this move?" AI explanation

---

## 5. Daily Cockpit Sidebar

### Calendar Agenda (top panel)
- Data source: Google Workspace CLI (`gws`) → Google Calendar API → local `.ics` → Cal.com
- Read-only: shows today + peek at tomorrow
- Color indicators for multiple calendars
- Refreshes every 15 minutes in background

### Task List (middle panel)
- Dead simple: title + done/not-done
- Keyboard: `a` add, `x` toggle, `d` delete, `j/k` navigate
- Persisted in SQLite
- No priorities, dates, or assignees — this is a scratchpad

### Standup Log (bottom panel)
- Three fields: Yesterday / Today / Blockers
- One entry per day, editable until end of day
- `[h]` scrollable history of past standups
- `[e]` inline editor for today's entry
- `[g]` AI-generated draft from git commits (optional)

---

## 6. AI Features & Graceful Degradation

### Core Principle
Every feature works without AI. AI makes it better, never required.

### Provider Chain (fallback order)
```
1. Ollama (local)     — always available, no network needed
2. Groq free tier     — fastest cloud, 1K-14K req/day
3. Gemini free tier   — most generous, 1K req/day
4. No-AI fallback     — app works without any AI
```

### AI Feature Map

| Feature | With AI | Without AI (fallback) |
|---------|---------|----------------------|
| News summary `[s]` | 2-3 sentence summary | First 200 chars of article |
| "Why did X move?" `[w]` | AI analyzes news + price action | List recent headlines |
| Standup draft `[g]` | AI cleans git commits into prose | Raw `git log --oneline` |
| News sentiment | AI-scored bullish/bearish/neutral | Keyword heuristic |
| Transaction categorization | Auto-tags sector | User picks from list |

### What AI Does NOT Do
- No trade recommendations ("should I buy?")
- No price predictions
- No automated alerts based on AI analysis
- No portfolio optimization suggestions

### Token Budget Awareness
- Track remaining quota per provider
- Status bar indicator: `AI: Groq ● 847/1000 req remaining`
- Auto-switch to next provider before hitting limits
- Cache AI responses (same prompt → cached result for 60min)

---

## 7. API Strategy

### Market Data

| Data | Primary API | Limits | Notes |
|------|------------|--------|-------|
| US Stocks/ETFs | Finnhub | 60 req/min (free) | Official Go SDK: `finnhub-go/v2`. WebSocket for real-time needs manual impl. |
| Crypto | CoinGecko | 5-15 req/min (free) | Go client: `goingecko`. Built-in rate limiting. |
| Currency | Frankfurter | Unlimited, no key | ECB data, 1 req/day sufficient |
| News (stocks) | Finnhub | Included in limits | Filtered by symbol |
| News (crypto) | CryptoPanic | 5 req/min (free) | Crypto-focused news |

### Portfolio Data
- **Primary:** Manual entry + Robinhood CSV import
- **Robinhood note:** No official stocks API. Crypto API exists (official). Stocks import via CSV export only.
- **Future:** SnapTrade or Plaid for brokerage aggregation

### Calendar
- **Priority order:** `auto` (try each in order)
  1. Google Workspace CLI (`gws calendar +agenda`) — simplest, zero OAuth code
  2. Direct Google Calendar API (Go client) — self-contained, no external dep
  3. Local `.ics` file — privacy-friendly, zero network
  4. Cal.com API — if user uses Cal.com

### AI Providers
- **Groq:** API key, sub-second latency, 1K-14K RPD depending on model
- **Gemini:** API key, 1K RPD, 250K TPM
- **Ollama:** Local, unlimited, OpenAI-compatible API on localhost:11434

### Smart Polling Strategy
- US market hours (9:30-16:00 ET): stocks every 30s
- Outside market hours: every 5 min
- Crypto: always every 60-120s (24/7 market)
- Calendar: every 15 min
- Currency: once daily
- Rate-limit-aware token bucket per API

---

## 8. Phase 2 Features (Post-MVP)

| Feature | Description |
|---------|-------------|
| Deep Research (NotebookLM) | `notebooklm-py` integration — feed SEC filings, earnings transcripts, create grounded Q&A per stock |
| Gmail via `gws` | Parse trade confirmation emails |
| Sheets import via `gws` | Import portfolio from Google Sheets |
| Post/Share | Screenshot portfolio to Twitter/Discord |
| International markets | Non-US exchanges |
| Cal.com integration | Alternative calendar source |
| Advanced charting | More technical indicators, drawing tools |
| Robinhood Crypto API | Real-time crypto position sync |
| Brokerage aggregation | SnapTrade/Plaid for multi-brokerage |

---

## 9. Success Criteria

A successful MVP should:
1. Display a real portfolio with live(ish) prices for US stocks and crypto
2. Render readable charts with at least one timeframe
3. Show today's calendar in the sidebar
4. Allow adding/completing tasks
5. Allow logging daily standup
6. Gracefully degrade when APIs are unavailable
7. Work as a single binary with no runtime dependencies
8. Be usable with Vim-style navigation
9. Look polished enough for a compelling README demo GIF
