package app

import "time"

// refreshTracker tracks when each data source was last refreshed.
type refreshTracker struct {
	lastRefresh map[string]time.Time
}

func newRefreshTracker() *refreshTracker {
	return &refreshTracker{lastRefresh: make(map[string]time.Time)}
}

// NeedsRefresh returns true if the given key hasn't been refreshed within maxAge.
func (r *refreshTracker) NeedsRefresh(key string, maxAge time.Duration) bool {
	last, ok := r.lastRefresh[key]
	return !ok || time.Since(last) > maxAge
}

// MarkRefreshed records the current time as the last refresh for key.
func (r *refreshTracker) MarkRefreshed(key string) {
	r.lastRefresh[key] = time.Now()
}
