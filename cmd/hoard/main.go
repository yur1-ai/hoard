package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/yur1-ai/hoard/internal/app"
	"github.com/yur1-ai/hoard/internal/config"
	"github.com/yur1-ai/hoard/internal/logger"
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

	model := app.New(cfg, db)
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
