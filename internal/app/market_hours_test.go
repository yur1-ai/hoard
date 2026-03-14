package app

import (
	"testing"
	"time"
)

func TestMarketOpenDuringTrading(t *testing.T) {
	// Tuesday 10:00 ET
	et := easternTZ
	ts := time.Date(2026, 3, 10, 10, 0, 0, 0, et) // Tuesday
	if !isUSMarketOpenAt(ts) {
		t.Error("expected market open at Tue 10:00 ET")
	}
}

func TestMarketClosedWeekend(t *testing.T) {
	et := easternTZ
	ts := time.Date(2026, 3, 14, 12, 0, 0, 0, et) // Saturday
	if isUSMarketOpenAt(ts) {
		t.Error("expected market closed on Saturday")
	}
}

func TestMarketClosedEvening(t *testing.T) {
	et := easternTZ
	ts := time.Date(2026, 3, 10, 20, 0, 0, 0, et) // Tuesday 8pm
	if isUSMarketOpenAt(ts) {
		t.Error("expected market closed at 20:00 ET")
	}
}

func TestMarketOpenAtBoundary(t *testing.T) {
	et := easternTZ
	// Exactly 9:30
	if !isUSMarketOpenAt(time.Date(2026, 3, 10, 9, 30, 0, 0, et)) {
		t.Error("expected open at 9:30")
	}
	// 9:29 — not open
	if isUSMarketOpenAt(time.Date(2026, 3, 10, 9, 29, 0, 0, et)) {
		t.Error("expected closed at 9:29")
	}
	// 16:00 — closed
	if isUSMarketOpenAt(time.Date(2026, 3, 10, 16, 0, 0, 0, et)) {
		t.Error("expected closed at 16:00")
	}
}

func TestExtendedHoursPreMarket(t *testing.T) {
	et := easternTZ
	ts := time.Date(2026, 3, 10, 7, 0, 0, 0, et) // Tuesday 7:00 pre-market
	if !isExtendedHoursAt(ts) {
		t.Error("expected extended hours at Tue 7:00 ET")
	}
}

func TestExtendedHoursPostMarket(t *testing.T) {
	et := easternTZ
	ts := time.Date(2026, 3, 10, 17, 0, 0, 0, et) // Tuesday 5pm post-market
	if !isExtendedHoursAt(ts) {
		t.Error("expected extended hours at Tue 17:00 ET")
	}
}

func TestExtendedHoursWeekend(t *testing.T) {
	et := easternTZ
	ts := time.Date(2026, 3, 14, 7, 0, 0, 0, et) // Saturday
	if isExtendedHoursAt(ts) {
		t.Error("expected no extended hours on Saturday")
	}
}
