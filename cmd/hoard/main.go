package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/yur1-ai/hoard/internal/app"
	"github.com/yur1-ai/hoard/internal/config"
	"github.com/yur1-ai/hoard/internal/logger"
	"github.com/yur1-ai/hoard/internal/service/currency"
	svcmarket "github.com/yur1-ai/hoard/internal/service/market"
	"github.com/yur1-ai/hoard/internal/store"
)

var version = "dev"

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
	// Check for subcommands before flag parsing (v3 fix)
	if len(os.Args) > 1 && os.Args[1] == "import" {
		runImport(os.Args[2:])
		return
	}

	flag.Parse()

	if flagVersion {
		fmt.Printf("hoard %s\n", version)
		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := logger.Init(config.LogFilePath(), flagDebug); err != nil {
		return fmt.Errorf("logger: %w", err)
	}

	var cfg config.Config
	var err error
	if flagConfig != "" {
		cfg, err = config.LoadFromFile(flagConfig)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if flagASCII {
		cfg.Theme = "ascii"
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation: %w", err)
	}

	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("dirs: %w", err)
	}

	db, err := store.Open(config.DBFilePath())
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer db.Close()

	// Auto-create default account on first run
	accounts, err := store.ListAccounts(db)
	if err != nil {
		return fmt.Errorf("list accounts: %w", err)
	}
	if len(accounts) == 0 {
		if _, err := store.CreateAccount(db, "Default", "brokerage", cfg.BaseCurrency); err != nil {
			return fmt.Errorf("create default account: %w", err)
		}
		slog.Info("created default account")
	}

	// Initialize market services
	stockTTL, err := time.ParseDuration(cfg.Market.RefreshIntervalMarket)
	if err != nil {
		return fmt.Errorf("parse stock refresh interval: %w", err)
	}
	cryptoTTL, err := time.ParseDuration(cfg.Market.RefreshIntervalCrypto)
	if err != nil {
		return fmt.Errorf("parse crypto refresh interval: %w", err)
	}

	var stockProvider svcmarket.StockProvider
	if cfg.Market.Finnhub.APIKey != "" {
		stockProvider = svcmarket.NewFinnhubClient(cfg.Market.Finnhub.APIKey, nil)
	}

	var cryptoProvider svcmarket.CryptoProvider = svcmarket.NewCoinGeckoClient(cfg.Market.CoinGecko.APIKey, nil)

	marketSvc := svcmarket.NewCachedService(stockProvider, cryptoProvider, db, stockTTL, cryptoTTL)
	currSvc := currency.NewFrankfurterClient(nil)

	model := app.New(cfg, db)
	model.SetMarketService(marketSvc)
	model.SetCurrencyService(currSvc)

	// Load holdings into portfolio view
	holdings, err := store.ListHoldings(db, 0) // all accounts
	if err != nil {
		slog.Warn("failed to load holdings", "error", err)
	} else {
		model.SetHoldings(holdings)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

// runImport handles: hoard import <format> <file>
// Stub for Phase 8 — prints usage for now.
func runImport(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: hoard import <format> <file>\nFormats: robinhood\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Import not yet implemented. Coming in a future release.\n")
	os.Exit(1)
}
