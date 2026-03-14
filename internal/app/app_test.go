package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/yur1-ai/hoard/internal/config"
)

func newTestApp() App {
	cfg := config.DefaultConfig()
	cfg.SidebarDefault = "closed"
	return New(cfg, nil)
}

func pressKey(m App, key string) App {
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: keyCode(key), Text: keyText(key)}))
	return updated.(App)
}

func keyCode(key string) rune {
	switch key {
	case "tab":
		return tea.KeyTab
	case "escape":
		return tea.KeyEscape
	default:
		if len(key) == 1 {
			return rune(key[0])
		}
		return 0
	}
}

func keyText(key string) string {
	if len(key) == 1 {
		return key
	}
	return ""
}

func TestViewSwitching(t *testing.T) {
	m := newTestApp()
	if m.activeView != viewPortfolio {
		t.Fatal("expected initial view = portfolio")
	}

	m = pressKey(m, "2")
	if m.activeView != viewWatchlist {
		t.Error("expected watchlist after pressing 2")
	}

	m = pressKey(m, "3")
	if m.activeView != viewCharts {
		t.Error("expected charts after pressing 3")
	}

	m = pressKey(m, "4")
	if m.activeView != viewNews {
		t.Error("expected news after pressing 4")
	}

	m = pressKey(m, "1")
	if m.activeView != viewPortfolio {
		t.Error("expected portfolio after pressing 1")
	}
}

func TestTabOpensClosedSidebar(t *testing.T) {
	m := newTestApp()
	m.sidebarOpen = false
	m.focus = focusMarket

	m = pressKey(m, "tab")
	if !m.sidebarOpen {
		t.Error("expected sidebar to open")
	}
	if m.focus != focusSidebar {
		t.Error("expected focus on sidebar")
	}
}

func TestTabClosesSidebarWhenFocused(t *testing.T) {
	m := newTestApp()
	m.sidebarOpen = true
	m.focus = focusSidebar

	m = pressKey(m, "tab")
	if m.sidebarOpen {
		t.Error("expected sidebar to close")
	}
	if m.focus != focusMarket {
		t.Error("expected focus on market")
	}
}

func TestTabFocusesSidebarWhenOpenAndMarketFocused(t *testing.T) {
	m := newTestApp()
	m.sidebarOpen = true
	m.focus = focusMarket

	m = pressKey(m, "tab")
	if !m.sidebarOpen {
		t.Error("expected sidebar to stay open")
	}
	if m.focus != focusSidebar {
		t.Error("expected focus on sidebar")
	}
}

func TestQuitKey(t *testing.T) {
	m := newTestApp()
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: rune('q'), Text: "q"}))
	if cmd == nil {
		t.Fatal("expected a quit command")
	}
	// Run the command and check for QuitMsg
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected QuitMsg, got %T", msg)
	}
}

func TestEscapeReturnsToNormalMode(t *testing.T) {
	m := newTestApp()
	m.mode = modeTextInput

	m = pressKey(m, "escape")
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal, got %d", m.mode)
	}
}

func TestSidebarDefaultOpen(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SidebarDefault = "open"
	m := New(cfg, nil)
	if !m.sidebarOpen {
		t.Error("expected sidebar open by default")
	}
}

func TestSidebarDefaultClosed(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SidebarDefault = "closed"
	m := New(cfg, nil)
	if m.sidebarOpen {
		t.Error("expected sidebar closed by default")
	}
}

func TestHelpToggle(t *testing.T) {
	m := newTestApp()
	if m.showHelp {
		t.Fatal("help should be off initially")
	}

	m = pressKey(m, "?")
	if !m.showHelp {
		t.Error("expected help on after ?")
	}

	// While help is showing, view keys should be blocked
	m = pressKey(m, "2")
	if m.activeView != viewPortfolio {
		t.Error("view should not change while help is open")
	}

	// ? again closes help
	m = pressKey(m, "?")
	if m.showHelp {
		t.Error("expected help off after second ?")
	}
}

func TestEscapeClosesHelp(t *testing.T) {
	m := newTestApp()
	m.showHelp = true

	m = pressKey(m, "escape")
	if m.showHelp {
		t.Error("escape should close help")
	}
}

func TestMinTerminalSize(t *testing.T) {
	m := newTestApp()

	// Simulate a small terminal
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	m = updated.(App)
	if !m.tooSmall {
		t.Error("expected tooSmall for 60x20")
	}

	// Simulate adequate terminal
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updated.(App)
	if m.tooSmall {
		t.Error("expected not tooSmall for 100x30")
	}
}

func TestViewSwitchingBlockedInTextInput(t *testing.T) {
	m := newTestApp()
	m.mode = modeTextInput

	m = pressKey(m, "2")
	if m.activeView != viewPortfolio {
		t.Error("view should not change in text input mode")
	}
}

func TestQuitBlockedInTextInput(t *testing.T) {
	m := newTestApp()
	m.mode = modeTextInput

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: rune('q'), Text: "q"}))
	if cmd != nil {
		t.Error("q should not quit in text input mode")
	}
}
