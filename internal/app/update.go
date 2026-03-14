package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"github.com/yur1-ai/hoard/internal/store"
	"github.com/yur1-ai/hoard/internal/ui/market/portfolio"
)

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tooSmall = msg.Width < minWidth || msg.Height < minHeight
		m = m.propagateSize()
		return m, nil

	case TickMsg:
		return m.handleTick()

	case QuotesMsg:
		m.portfolio.UpdateQuotes(msg.Quotes)
		m = m.updateHeaderFromPortfolio()
		return m, nil

	case CryptoQuotesMsg:
		m.portfolio.UpdateQuotes(msg.Quotes)
		m = m.updateHeaderFromPortfolio()
		return m, nil

	case CurrencyRatesMsg:
		slog.Info("currency rates updated", "count", len(msg.Rates))
		return m, nil

	case ErrMsg:
		m.lastErr = msg.Error()
		return m, nil

	case portfolio.AddHoldingMsg:
		return m.handleAddHolding(msg)

	case portfolio.CancelFormMsg:
		m.portfolio.HideForm()
		m.mode = modeNormal
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	default:
		// Forward unhandled messages (e.g. textinput cursor blink) when form is active
		if m.mode == modeTextInput {
			var cmd tea.Cmd
			m.portfolio, cmd = m.portfolio.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m App) handleTick() (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	cmds = append(cmds, m.scheduleTick())

	if m.marketSvc == nil || m.db == nil {
		return m, tea.Batch(cmds...)
	}

	// Equity refresh — adaptive interval based on market hours
	equityInterval := 5 * time.Minute
	if isUSMarketOpen() {
		equityInterval = 30 * time.Second
	} else if isExtendedHours() {
		equityInterval = 2 * time.Minute
	}
	if m.refresh.NeedsRefresh("equity", equityInterval) {
		symbols, err := store.AllEquitySymbols(m.db)
		if err != nil {
			slog.Warn("failed to get equity symbols", "error", err)
		} else if len(symbols) > 0 {
			svc := m.marketSvc
			cmds = append(cmds, safeCmd(func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				quotes, err := svc.GetStockQuotes(ctx, symbols)
				if err != nil {
					return ErrMsg{Err: err, Context: "stock quotes"}
				}
				return QuotesMsg{Quotes: quotes}
			}))
		}
		m.refresh.MarkRefreshed("equity")
	}

	// Crypto refresh (24/7, every 120s)
	if m.refresh.NeedsRefresh("crypto", 120*time.Second) {
		symbols, err := store.AllCryptoSymbols(m.db)
		if err != nil {
			slog.Warn("failed to get crypto symbols", "error", err)
		} else if len(symbols) > 0 {
			svc := m.marketSvc
			cmds = append(cmds, safeCmd(func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				quotes, err := svc.GetCryptoQuotes(ctx, symbols)
				if err != nil {
					return ErrMsg{Err: err, Context: "crypto quotes"}
				}
				return CryptoQuotesMsg{Quotes: quotes}
			}))
		}
		m.refresh.MarkRefreshed("crypto")
	}

	// Currency rates (daily)
	if m.refresh.NeedsRefresh("currency", 24*time.Hour) && m.currSvc != nil {
		currSvc := m.currSvc
		db := m.db
		base := m.cfg.BaseCurrency
		cmds = append(cmds, safeCmd(func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			rates, err := currSvc.FetchRates(ctx, base)
			if err != nil {
				return ErrMsg{Err: fmt.Errorf("currency rates: %w", err), Context: "currency"}
			}
			for to, rate := range rates {
				if err := store.SetCurrencyRate(db, base, to, rate); err != nil {
					slog.Warn("failed to cache currency rate", "pair", base+"→"+to, "error", err)
				}
			}
			return CurrencyRatesMsg{Rates: rates}
		}))
		m.refresh.MarkRefreshed("currency")
	}

	return m, tea.Batch(cmds...)
}

func (m App) updateHeaderFromPortfolio() App {
	m.headerData.TotalValue = m.portfolio.TotalValue()
	m.headerData.DayChange = m.portfolio.DayChange()
	m.headerData.DayPct = m.portfolio.DayChangePct()
	return m
}

func (m App) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Help overlay: only ? and Esc work while showing
	if m.showHelp {
		if key.Matches(msg, m.keys.Help) || key.Matches(msg, m.keys.Escape) {
			m.showHelp = false
		}
		return m, nil
	}

	// Escape always resets to normal mode and closes overlays
	if key.Matches(msg, m.keys.Escape) {
		m.portfolio.HideForm()
		m.mode = modeNormal
		m.lastErr = ""
		return m, nil
	}

	// In text input mode, forward all non-Esc keys to the active sub-model
	if m.mode == modeTextInput {
		var cmd tea.Cmd
		m.portfolio, cmd = m.portfolio.Update(msg)
		return m, cmd
	}

	// Global keys (normal mode)
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp

	case key.Matches(msg, m.keys.View1):
		m.activeView = viewPortfolio
	case key.Matches(msg, m.keys.View2):
		m.activeView = viewWatchlist
	case key.Matches(msg, m.keys.View3):
		m.activeView = viewCharts
	case key.Matches(msg, m.keys.View4):
		m.activeView = viewNews

	case key.Matches(msg, m.keys.Tab):
		m = m.toggleSidebar()

	case key.Matches(msg, m.keys.Add):
		if m.activeView == viewPortfolio {
			cmd := m.portfolio.ShowForm()
			m.mode = modeTextInput
			return m, cmd
		}

	case key.Matches(msg, m.keys.Delete):
		if m.activeView == viewPortfolio {
			return m.handleDeleteHolding()
		}

	case key.Matches(msg, m.keys.Refresh):
		m.refresh.Reset("equity")
		m.refresh.Reset("crypto")
		return m.handleTick()

	default:
		// Route to focused area for navigation keys (j/k/up/down)
		if m.focus == focusSidebar {
			var cmd tea.Cmd
			m.sidebar, cmd = m.sidebar.Update(msg)
			return m, cmd
		}
		if m.activeView == viewPortfolio {
			var cmd tea.Cmd
			m.portfolio, cmd = m.portfolio.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m App) handleAddHolding(msg portfolio.AddHoldingMsg) (tea.Model, tea.Cmd) {
	m.portfolio.HideForm()
	m.mode = modeNormal

	if m.db == nil {
		return m, nil
	}

	// Get default account (first available)
	accounts, err := store.ListAccounts(m.db)
	if err != nil || len(accounts) == 0 {
		m.lastErr = "no account available"
		return m, nil
	}

	// Record buy transaction (auto-creates/updates holding via applyBuy)
	_, err = store.AddTransaction(m.db, store.Transaction{
		AccountID: accounts[0].ID,
		Symbol:    msg.Symbol,
		Market:    msg.Market,
		Type:      "buy",
		Quantity:  msg.Quantity,
		Price:     msg.AvgCost,
		Date:      time.Now(),
	})
	if err != nil {
		m.lastErr = fmt.Sprintf("add position: %v", err)
		return m, nil
	}

	return m.reloadHoldings()
}

func (m App) handleDeleteHolding() (tea.Model, tea.Cmd) {
	if m.db == nil {
		return m, nil
	}
	h := m.portfolio.SelectedHolding()
	if h == nil {
		return m, nil
	}
	if err := store.DeleteHolding(m.db, h.ID); err != nil {
		m.lastErr = fmt.Sprintf("delete: %v", err)
		return m, nil
	}
	return m.reloadHoldings()
}

func (m App) reloadHoldings() (tea.Model, tea.Cmd) {
	holdings, err := store.ListHoldings(m.db, 0)
	if err != nil {
		m.lastErr = fmt.Sprintf("reload: %v", err)
		return m, nil
	}
	m.portfolio.SetHoldings(holdings)
	m = m.updateHeaderFromPortfolio()

	// Trigger immediate refresh to fetch prices for any new symbols
	m.refresh.Reset("equity")
	m.refresh.Reset("crypto")
	return m.handleTick()
}

func (m App) toggleSidebar() App {
	switch {
	case !m.sidebarOpen:
		m.sidebarOpen = true
		m.focus = focusSidebar
	case m.focus == focusMarket:
		m.focus = focusSidebar
	default:
		m.sidebarOpen = false
		m.focus = focusMarket
	}
	m.sidebar.Focused = m.focus == focusSidebar
	return m.propagateSize()
}

func (m App) propagateSize() App {
	marketWidth := m.width
	contentHeight := m.height - 2 // header + footer

	if m.sidebarOpen {
		marketWidth -= m.sidebar.Width
	}
	if marketWidth < 0 {
		marketWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	m.sidebar.Height = contentHeight

	m.portfolio.SetSize(marketWidth, contentHeight)
	m.watchlist.SetSize(marketWidth, contentHeight)
	m.charts.SetSize(marketWidth, contentHeight)
	m.news.SetSize(marketWidth, contentHeight)
	return m
}
