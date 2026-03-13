package app

import "time"

// TickMsg fires on each refresh interval.
type TickMsg time.Time

// ErrMsg carries a non-fatal error to display in the status bar.
type ErrMsg struct {
	Err     error
	Context string
}

func (e ErrMsg) Error() string { return e.Err.Error() }
