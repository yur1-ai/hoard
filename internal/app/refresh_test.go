package app

import (
	"testing"
	"time"
)

func TestRefreshTrackerNeedsRefresh(t *testing.T) {
	rt := newRefreshTracker()

	// Never refreshed → needs refresh
	if !rt.NeedsRefresh("equity", 30*time.Second) {
		t.Error("expected NeedsRefresh for never-refreshed key")
	}

	// Just refreshed → doesn't need refresh
	rt.MarkRefreshed("equity")
	if rt.NeedsRefresh("equity", 30*time.Second) {
		t.Error("expected no refresh needed right after mark")
	}

	// Different key → still needs refresh
	if !rt.NeedsRefresh("crypto", 120*time.Second) {
		t.Error("expected NeedsRefresh for different key")
	}
}

func TestRefreshTrackerExpiry(t *testing.T) {
	rt := newRefreshTracker()
	rt.lastRefresh["equity"] = time.Now().Add(-2 * time.Minute)

	if !rt.NeedsRefresh("equity", 30*time.Second) {
		t.Error("expected NeedsRefresh for stale key")
	}
}
