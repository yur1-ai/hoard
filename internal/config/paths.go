package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns ~/.config/hoard (or XDG_CONFIG_HOME/hoard).
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "hoard")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "hoard")
}

// DataDir returns ~/.local/share/hoard (or XDG_DATA_HOME/hoard).
func DataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "hoard")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "hoard")
}

// ConfigFilePath returns the full path to config.toml.
func ConfigFilePath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

// DBFilePath returns the full path to hoard.db.
func DBFilePath() string {
	return filepath.Join(DataDir(), "hoard.db")
}

// LogFilePath returns the full path to hoard.log.
func LogFilePath() string {
	return filepath.Join(DataDir(), "hoard.log")
}

// EnsureDirs creates config and data directories if they don't exist.
func EnsureDirs() error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}
	return os.MkdirAll(DataDir(), 0755)
}
