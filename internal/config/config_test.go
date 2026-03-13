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

func TestCoinGeckoEnvOverride(t *testing.T) {
	t.Setenv("HOARD_COINGECKO_KEY", "cg-demo-key")
	cfg := DefaultConfig()
	cfg.ApplyEnvOverrides()
	if cfg.Market.CoinGecko.APIKey != "cg-demo-key" {
		t.Errorf("expected cg-demo-key, got %s", cfg.Market.CoinGecko.APIKey)
	}
}

func TestValidateAcceptsDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("defaults should validate, got: %v", err)
	}
}

func TestValidateRejectsInvalidTheme(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Theme = "rainbow"
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid theme")
	}
}

func TestValidateRejectsInvalidSidebar(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SidebarDefault = "maybe"
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid sidebar_default")
	}
}

func TestValidateRejectsInvalidRefreshInterval(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Market.RefreshIntervalMarket = "not-a-duration"
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for bad duration")
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with missing file should not error, got: %v", err)
	}
	if cfg.BaseCurrency != "USD" {
		t.Errorf("expected defaults, got currency %s", cfg.BaseCurrency)
	}
}
