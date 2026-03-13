package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/yurishevtsov/hoard/internal/app"
	"github.com/yurishevtsov/hoard/internal/config"
	"github.com/yurishevtsov/hoard/internal/store"
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
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
