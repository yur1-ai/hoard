package app

import (
	"database/sql"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/yur1-ai/hoard/internal/config"
	"github.com/yur1-ai/hoard/internal/service/currency"
	svcmarket "github.com/yur1-ai/hoard/internal/service/market"
	"github.com/yur1-ai/hoard/internal/store"
	"github.com/yur1-ai/hoard/internal/ui/common"
	"github.com/yur1-ai/hoard/internal/ui/header"
	"github.com/yur1-ai/hoard/internal/ui/market/charts"
	"github.com/yur1-ai/hoard/internal/ui/market/news"
	"github.com/yur1-ai/hoard/internal/ui/market/portfolio"
	"github.com/yur1-ai/hoard/internal/ui/market/watchlist"
	"github.com/yur1-ai/hoard/internal/ui/sidebar"
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

const (
	minWidth  = 80
	minHeight = 24
)

// App is the root Bubble Tea model.
type App struct {
	cfg    config.Config
	db     *sql.DB
	keys   common.KeyMap
	width  int
	height int

	mode        inputMode
	activeView  activeView
	focus       focusArea
	sidebarOpen bool
	showHelp    bool
	tooSmall    bool
	lastErr     string

	marketSvc *svcmarket.CachedService
	currSvc   *currency.FrankfurterClient
	refresh   *refreshTracker

	headerData header.Data
	sidebar    sidebar.Model
	portfolio  portfolio.Model
	watchlist  watchlist.Model
	charts     charts.Model
	news       news.Model
}

func New(cfg config.Config, db *sql.DB) App {
	sidebarOpen := cfg.SidebarDefault == "open"
	return App{
		cfg:         cfg,
		db:          db,
		keys:        common.DefaultKeyMap(),
		mode:        modeNormal,
		activeView:  viewPortfolio,
		focus:       focusMarket,
		sidebarOpen: sidebarOpen,
		refresh:     newRefreshTracker(),
		sidebar:     sidebar.New(),
		portfolio:   portfolio.New(),
		watchlist:   watchlist.New(),
		charts:      charts.New(),
		news:        news.New(),
	}
}

// SetMarketService injects the market data service (called from main before tea.Program.Run).
func (m *App) SetMarketService(svc *svcmarket.CachedService) {
	m.marketSvc = svc
}

// SetCurrencyService injects the currency rate service.
func (m *App) SetCurrencyService(svc *currency.FrankfurterClient) {
	m.currSvc = svc
}

// SetHoldings loads initial holdings into the portfolio view.
func (m *App) SetHoldings(holdings []store.Holding) {
	m.portfolio.SetHoldings(holdings)
}

const tickInterval = 10 * time.Second

func (m App) Init() tea.Cmd {
	return m.scheduleTick()
}

func (m App) scheduleTick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
