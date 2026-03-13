# Hoard — Research Findings

**Date:** 2026-03-13

---

## 1. TUI Ecosystem Landscape

### Dominant TUI Apps (for context/inspiration)

| App | Stars | Language | What It Does |
|-----|-------|----------|-------------|
| lazygit | 74.1k | Go | Git client |
| rich | 55.8k | Python | Terminal formatting library |
| dive | 53.6k | Go | Docker image explorer |
| bubbletea | 40.6k | Go | TUI framework |
| textual | 34.8k | Python | TUI framework |
| yazi | 33.9k | Rust | File manager |
| ratatui | 19k | Rust | TUI framework |
| posting | 11.5k | Python | API client |
| ticker | 6k | Go | Stock price ticker |
| Bagels | 2.7k | Python | Expense tracker |

### What Makes TUI Apps Go Viral
1. **"Lazygit formula"** — wrap complex CLI in beautiful keyboard-driven UI
2. **Single SQLite file** — local-first, no account needed
3. **Vim keybindings** — table stakes
4. **Gorgeous demo GIF** in README (use VHS for recording)
5. **"Replaces an Electron app"** framing
6. **Works over SSH** — unique advantage over GUI

### Gap Analysis

| Category | Best TUI Option | Gap Size |
|----------|----------------|----------|
| Portfolio tracking | ticker (view-only) | MASSIVE |
| Personal finance | Bagels (expenses only) | MASSIVE |
| Knowledge base | Nearly nothing | MASSIVE |
| Kanban/project mgmt | taskwarrior-tui | MASSIVE |
| Calendar | khal (12 years old) | LARGE |
| Work journal | tui-journal (basic) | LARGE |
| Messaging | gomuks (basic Matrix) | LARGE |

---

## 2. Free AI APIs

### Permanently Free (best for Hoard)

| Provider | Best Model | Limits | Latency | Notes |
|----------|-----------|--------|---------|-------|
| **Groq** | Llama 3.3 70B | 1K-14.4K RPD | <1s | Fastest cloud inference |
| **Google Gemini** | Flash-Lite | 1K RPD, 250K TPM | 1-3s | Most generous free tier |
| **Mistral** | All models | 1B tokens/month | 1-3s | Data used for training on free tier |
| **Ollama** (local) | Llama 3.3, Qwen3 | Unlimited | 0.5-5s | Requires 8GB+ RAM |
| **OpenRouter** | 28+ free models | 50 RPD | 2-5s | OpenAI-compatible endpoint |

### Recommended Chain for Hoard
```
1. Ollama (local)  → always available, zero latency
2. Groq            → fastest cloud, generous limits
3. Gemini          → high quality, decent limits
4. No-AI fallback  → everything still works
```

---

## 3. TUI Framework Decision: Go + Bubble Tea v2

### Why Bubble Tea Over Alternatives

| Factor | Bubble Tea (Go) | Ratatui (Rust) | Textual (Python) | Ink (JS/React) |
|--------|----------------|----------------|-------------------|----------------|
| Dev velocity | Fast | Slow | Fastest | Fast |
| AI streaming | Natural (goroutines) | Manual (tokio) | Natural (asyncio) | Natural (React) |
| Binary dist | Single binary | Single binary | Needs Python | Needs Node.js |
| Learning curve | Moderate | Steep | Low | Low (if React) |
| Performance | Very good (v2) | Best | Adequate | Good |

### Bubble Tea v2 Key Changes (Feb 2026)
- Import paths: `charm.land/bubbletea/v2` (not `github.com/...`)
- `View()` returns `tea.View` struct (not `string`)
- `tea.KeyMsg` → `tea.KeyPressMsg` + `tea.KeyReleaseMsg`
- New ncurses-based renderer: ~10x faster
- Color auto-downsampling built-in
- Mouse events split into Click/Release/Wheel/Motion

---

## 4. API-Specific Research

### Finnhub (Stocks/ETFs)
- **Official Go SDK:** `github.com/Finnhub-Stock-API/finnhub-go/v2`
- **Free tier:** 60 req/min
- **Coverage:** Quotes, candles, fundamentals, filings, news, sentiment, technicals
- **WebSocket:** NOT in Go SDK — manual implementation needed for real-time
- **Docs:** https://finnhub.io/docs/api

### CoinGecko (Crypto)
- **Go client:** `github.com/JulianToledano/goingecko` v3 (active, maintained)
- **WARNING:** `go-gecko` (superoo7) is ARCHIVED — do not use
- **Free tier:** 5-15 req/min (more restrictive than initially expected)
- **Coverage:** Coins, markets, OHLC, exchanges, categories
- **Built-in:** Rate limiting, exponential backoff, retry logic

### Frankfurter (Currency)
- **No SDK needed** — simple REST: `https://api.frankfurter.dev/v1/latest?base=USD`
- **Limits:** None. No API key.
- **Source:** European Central Bank data
- **Refresh:** Once daily is sufficient

### Robinhood
- **Official API:** Crypto Trading API only (documented, supported)
- **Stocks API:** NO official public API for retail developers
- **Unofficial APIs:** Exist but violate ToS, risk account suspension
- **Practical approach:** CSV import for stocks, official API for crypto only
- **Source:** https://docs.robinhood.com/

### Google Workspace CLI (`gws`)
- **Released:** March 2026 (brand new, open-source Apache 2.0)
- **Repo:** https://github.com/googleworkspace/cli
- **Install:** `npm install -g @googleworkspace/cli` or pre-built binaries
- **Calendar command:** `gws calendar +agenda`
- **Coverage:** Drive, Gmail, Calendar, Sheets, Docs, Chat, Admin
- **Auth:** `gws auth setup` + `gws auth login` (OAuth, one-time)
- **Future value:** Gmail for trade confirmations, Sheets for portfolio import
- **Caveat:** "Not an officially supported Google product"

### notebooklm-py (Phase 2)
- **Repo:** https://github.com/teng-lin/notebooklm-py
- **What:** Unofficial Python API + CLI for Google NotebookLM
- **Features:** Add sources (URLs, PDFs, YouTube), grounded Q&A, generate podcasts/reports/flashcards
- **Auth:** Browser-based Google login
- **Requires:** Python 3.10+, Playwright (for auth)
- **Use case for Hoard:** "Deep Research" mode — feed SEC filings + earnings transcripts, ask grounded questions per stock
- **Risk:** Unofficial APIs, could break anytime

---

## 5. SQLite (modernc.org/sqlite)

### Why modernc over mattn/go-sqlite3
- **Pure Go:** No CGO, no C compiler needed
- **Cross-compilation:** Single `go build` for all platforms
- **Performance:** Read-heavy queries nearly identical (~130ms vs ~120ms)
- **Write penalty:** 2-3x slower on bulk inserts (irrelevant for our use case)
- **Production-ready:** Used by Gogs and other production projects

### Gotchas
- Pin `modernc.org/libc` version exactly — mismatches cause build failures
- WAL mode: always `defer rows.Close()` to prevent WAL file bloat
- Periodic `PRAGMA wal_checkpoint(TRUNCATE)` to manage WAL size
- Embed migrations with `go:embed` for single-binary distribution

---

## 6. Terminal Chart Libraries

### Primary: NimbleMarkets/ntcharts
- **Stars:** 664 (as of March 2026)
- **Chart types:** Line, candlestick, bar, sparkline, scatter, heatmap, canvas
- **Built for Bubble Tea** — returns string from View()
- **Risk:** Currently imports Bubble Tea v1. May need v2 migration or fork.

### Fallback: asciigraph
- **Stars:** 2.7k
- **Zero dependencies** — renders line charts as plain strings
- **Limited but reliable** — no Bubble Tea coupling

### Alternative: Custom sparklines
- Sparkline rendering is ~50 lines with braille characters
- Good enough for inline portfolio sparklines

---

## 7. Common Bubble Tea Pitfalls

1. **Never mutate model outside Update()** — commands return messages, Update processes them
2. **Never block the event loop** — no `time.Sleep`, no sync HTTP, no DB queries in Update/View
3. **Command ordering is non-deterministic** — `tea.Batch` runs concurrently, use `tea.Sequence` if order matters
4. **Panics in commands don't reset terminal** — wrap with `recover()`
5. **Forward `WindowSizeMsg` to all nested models** — or panels won't resize
6. **Use `context.WithTimeout` on all HTTP calls** — no built-in timeout in tea.Cmd
7. **Check input mode before processing shortcuts** — or typing "1" in a text field switches views

---

## 8. Sources

### TUI Ecosystem
- [awesome-tuis (rothgar)](https://github.com/rothgar/awesome-tuis)
- [awesome-ratatui](https://github.com/ratatui/awesome-ratatui)
- [Terminal Trove](https://terminaltrove.com/)

### Bubble Tea
- [Bubble Tea v2 Upgrade Guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md)
- [Charm v2 Blog Post](https://charm.land/blog/v2/)
- [Commands in Bubble Tea](https://charm.land/blog/commands-in-bubbletea/)
- [Tips for building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/)
- [Loss of input in Bubbletea (critique)](https://dr-knz.net/bubbletea-control-inversion.html)
- [Testing Bubble Tea interfaces](https://patternmatched.substack.com/p/testing-bubble-tea-interfaces)

### AI APIs
- [Free LLM API Resources (GitHub)](https://github.com/cheahjs/free-llm-api-resources)
- [Groq Rate Limits](https://console.groq.com/docs/rate-limits)
- [Gemini Rate Limits](https://ai.google.dev/gemini-api/docs/rate-limits)

### Market APIs
- [Finnhub API Docs](https://finnhub.io/docs/api)
- [CoinGecko API Docs](https://docs.coingecko.com/)
- [Frankfurter API](https://api.frankfurter.dev/)

### Tools
- [Google Workspace CLI](https://github.com/googleworkspace/cli)
- [notebooklm-py](https://github.com/teng-lin/notebooklm-py)
- [Robinhood API Docs](https://docs.robinhood.com/)
- [go-sqlite-bench](https://github.com/cvilsmeier/go-sqlite-bench)
