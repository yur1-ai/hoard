package config

import (
	"fmt"
	"os"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	BaseCurrency   string `toml:"base_currency"`
	Theme          string `toml:"theme"`
	SidebarDefault string `toml:"sidebar_default"`

	Market   MarketConfig   `toml:"market"`
	Calendar CalendarConfig `toml:"calendar"`
	AI       AIConfig       `toml:"ai"`
}

type MarketConfig struct {
	StockProvider         string          `toml:"stock_provider"`
	CryptoProvider        string          `toml:"crypto_provider"`
	RefreshIntervalMarket string          `toml:"refresh_interval_market"`
	RefreshIntervalCrypto string          `toml:"refresh_interval_crypto"`
	RefreshIntervalClosed string          `toml:"refresh_interval_closed"`
	Finnhub               FinnhubConfig   `toml:"finnhub"`
	CoinGecko             CoinGeckoConfig `toml:"coingecko"`
}

type FinnhubConfig struct {
	APIKey string `toml:"api_key"`
}

type CoinGeckoConfig struct {
	APIKey string `toml:"api_key"`
}

type CalendarConfig struct {
	Source  string `toml:"source"`
	ICSPath string `toml:"ics_path"`
}

type AIConfig struct {
	ProviderPriority []string     `toml:"provider_priority"`
	TimeoutMs        int          `toml:"timeout_ms"`
	CacheTTLMinutes  int          `toml:"cache_ttl_minutes"`
	Ollama           OllamaConfig `toml:"ollama"`
	Groq             GroqConfig   `toml:"groq"`
	Gemini           GeminiConfig `toml:"gemini"`
}

type OllamaConfig struct {
	Endpoint string `toml:"endpoint"`
	Model    string `toml:"model"`
}

type GroqConfig struct {
	APIKey string `toml:"api_key"`
}

type GeminiConfig struct {
	APIKey string `toml:"api_key"`
}

func DefaultConfig() Config {
	return Config{
		BaseCurrency:   "USD",
		Theme:          "dark",
		SidebarDefault: "open",
		Market: MarketConfig{
			StockProvider:         "finnhub",
			CryptoProvider:        "coingecko",
			RefreshIntervalMarket: "30s",
			RefreshIntervalCrypto: "120s",
			RefreshIntervalClosed: "5m",
		},
		Calendar: CalendarConfig{
			Source: "auto",
		},
		AI: AIConfig{
			ProviderPriority: []string{"ollama", "groq", "gemini"},
			TimeoutMs:        5000,
			CacheTTLMinutes:  60,
			Ollama: OllamaConfig{
				Endpoint: "http://localhost:11434",
				Model:    "llama3.3",
			},
		},
	}
}

func LoadFromFile(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	cfg.ApplyEnvOverrides()
	return cfg, nil
}

// Load tries the default config path; returns defaults if file missing.
func Load() (Config, error) {
	path := ConfigFilePath()
	cfg, err := LoadFromFile(path)
	if os.IsNotExist(err) {
		cfg = DefaultConfig()
		cfg.ApplyEnvOverrides()
		return cfg, nil
	}
	return cfg, err
}

func (c *Config) ApplyEnvOverrides() {
	if v := os.Getenv("HOARD_FINNHUB_KEY"); v != "" {
		c.Market.Finnhub.APIKey = v
	}
	if v := os.Getenv("HOARD_COINGECKO_KEY"); v != "" {
		c.Market.CoinGecko.APIKey = v
	}
	if v := os.Getenv("HOARD_GROQ_KEY"); v != "" {
		c.AI.Groq.APIKey = v
	}
	if v := os.Getenv("HOARD_GEMINI_KEY"); v != "" {
		c.AI.Gemini.APIKey = v
	}
}

func (c Config) Validate() error {
	validThemes := map[string]bool{"dark": true, "light": true, "auto": true, "ascii": true}
	if !validThemes[c.Theme] {
		return fmt.Errorf("invalid theme %q: must be dark, light, auto, or ascii", c.Theme)
	}
	validSidebar := map[string]bool{"open": true, "closed": true}
	if !validSidebar[c.SidebarDefault] {
		return fmt.Errorf("invalid sidebar_default %q: must be open or closed", c.SidebarDefault)
	}
	if _, err := time.ParseDuration(c.Market.RefreshIntervalMarket); err != nil {
		return fmt.Errorf("invalid refresh_interval_market: %w", err)
	}
	if _, err := time.ParseDuration(c.Market.RefreshIntervalCrypto); err != nil {
		return fmt.Errorf("invalid refresh_interval_crypto: %w", err)
	}
	if _, err := time.ParseDuration(c.Market.RefreshIntervalClosed); err != nil {
		return fmt.Errorf("invalid refresh_interval_closed: %w", err)
	}
	return nil
}

// WriteConfigFile writes config data with restrictive permissions (API keys inside).
func WriteConfigFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0600)
}
