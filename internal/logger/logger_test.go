package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCreatesLogFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	if err := Init(logPath, false); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	slog.Info("test message", "key", "value")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if len(data) == 0 {
		t.Error("log file is empty after writing")
	}
}

func TestInitDebugLevel(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "debug.log")

	if err := Init(logPath, true); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	slog.Debug("debug message")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if len(data) == 0 {
		t.Error("debug message not written when debug=true")
	}
}

func TestDiscardDoesNotPanic(t *testing.T) {
	Discard()
	slog.Info("should not panic")
	slog.Error("should not panic either")
}
