package app

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/yurishevtsov/hoard/internal/logger"
)

func init() {
	logger.Discard()
}

func TestSafeCmdRecoversPanic(t *testing.T) {
	panicCmd := safeCmd(func() tea.Msg {
		panic("test panic")
	})
	msg := panicCmd() // should NOT panic
	errMsg, ok := msg.(ErrMsg)
	if !ok {
		t.Fatalf("expected ErrMsg from recovered panic, got %T", msg)
	}
	if errMsg.Context != "panic" {
		t.Errorf("expected panic context, got %s", errMsg.Context)
	}
}

func TestSafeCmdPassesThrough(t *testing.T) {
	normalCmd := safeCmd(func() tea.Msg {
		return TickMsg(time.Now())
	})
	msg := normalCmd()
	if _, ok := msg.(TickMsg); !ok {
		t.Fatalf("expected TickMsg passthrough, got %T", msg)
	}
}

func TestSafeCmdHandlesNilReturn(t *testing.T) {
	nilCmd := safeCmd(func() tea.Msg {
		return nil
	})
	msg := nilCmd()
	if msg != nil {
		t.Errorf("expected nil passthrough, got %T", msg)
	}
}
