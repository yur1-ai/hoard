package app

import (
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"
)

// safeCmd wraps a tea.Cmd function with panic recovery.
// If the function panics, it returns an ErrMsg instead of crashing the terminal.
func safeCmd(fn func() tea.Msg) tea.Cmd {
	return func() (result tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in command", "recover", r)
				result = ErrMsg{
					Err:     fmt.Errorf("internal error: %v", r),
					Context: "panic",
				}
			}
		}()
		return fn()
	}
}
