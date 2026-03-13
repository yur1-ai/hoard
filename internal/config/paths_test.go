package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDirXDGOverride(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg-config")
	got := ConfigDir()
	want := "/tmp/test-xdg-config/hoard"
	if got != want {
		t.Errorf("ConfigDir() = %s, want %s", got, want)
	}
}

func TestDataDirXDGOverride(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/test-xdg-data")
	got := DataDir()
	want := "/tmp/test-xdg-data/hoard"
	if got != want {
		t.Errorf("DataDir() = %s, want %s", got, want)
	}
}

func TestConfigDirDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	got := ConfigDir()
	want := filepath.Join(home, ".config", "hoard")
	if got != want {
		t.Errorf("ConfigDir() = %s, want %s", got, want)
	}
}

func TestEnsureDirs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))

	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs() error: %v", err)
	}

	for _, path := range []string{ConfigDir(), DataDir()} {
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("directory %s not created: %v", path, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", path)
		}
	}
}
