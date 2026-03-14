package portfolio

import (
	"math"
	"testing"

	"github.com/yur1-ai/hoard/internal/service/market"
	"github.com/yur1-ai/hoard/internal/store"
)

func approxEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestCalculatePnL(t *testing.T) {
	m := New()
	m.width = 120
	m.height = 30
	m.SetHoldings([]store.Holding{
		{Symbol: "AAPL", Quantity: 10, AvgCostBasis: 150.0},
		{Symbol: "GOOG", Quantity: 5, AvgCostBasis: 130.0},
	})

	m.UpdateQuotes([]market.Quote{
		{Symbol: "AAPL", Price: 175.0, Change: 2.0, ChangePct: 1.15},
		{Symbol: "GOOG", Price: 140.0, Change: -1.0, ChangePct: -0.71},
	})

	// AAPL: P&L = 10 * (175 - 150) = 250
	if !approxEqual(m.holdings[0].pnl, 250.0, 0.01) {
		t.Errorf("AAPL P&L: expected 250, got %f", m.holdings[0].pnl)
	}

	// GOOG: P&L = 5 * (140 - 130) = 50
	if !approxEqual(m.holdings[1].pnl, 50.0, 0.01) {
		t.Errorf("GOOG P&L: expected 50, got %f", m.holdings[1].pnl)
	}

	// Total value: (10 * 175) + (5 * 140) = 1750 + 700 = 2450
	if !approxEqual(m.TotalValue(), 2450.0, 0.01) {
		t.Errorf("total value: expected 2450, got %f", m.TotalValue())
	}

	// Day change: (10 * 2) + (5 * -1) = 20 - 5 = 15
	if !approxEqual(m.DayChange(), 15.0, 0.01) {
		t.Errorf("day change: expected 15, got %f", m.DayChange())
	}
}

func TestCalculateAllocation(t *testing.T) {
	m := New()
	m.width = 120
	m.height = 30
	m.SetHoldings([]store.Holding{
		{Symbol: "AAPL", Quantity: 10, AvgCostBasis: 100.0},
		{Symbol: "GOOG", Quantity: 10, AvgCostBasis: 100.0},
	})

	m.UpdateQuotes([]market.Quote{
		{Symbol: "AAPL", Price: 100.0},
		{Symbol: "GOOG", Price: 100.0},
	})

	// Equal allocation: 50% each
	if !approxEqual(m.holdings[0].allocPct, 50.0, 0.1) {
		t.Errorf("AAPL alloc: expected 50%%, got %f%%", m.holdings[0].allocPct)
	}
	if !approxEqual(m.holdings[1].allocPct, 50.0, 0.1) {
		t.Errorf("GOOG alloc: expected 50%%, got %f%%", m.holdings[1].allocPct)
	}
}

func TestEmptyPortfolioView(t *testing.T) {
	m := New()
	m.SetSize(120, 30)

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view for empty portfolio")
	}
}

func TestPnLPercentage(t *testing.T) {
	m := New()
	m.width = 120
	m.height = 30
	m.SetHoldings([]store.Holding{
		{Symbol: "AAPL", Quantity: 10, AvgCostBasis: 100.0},
	})

	m.UpdateQuotes([]market.Quote{
		{Symbol: "AAPL", Price: 120.0},
	})

	// P&L% = (120 - 100) / 100 * 100 = 20%
	if !approxEqual(m.holdings[0].pnlPct, 20.0, 0.01) {
		t.Errorf("expected P&L%% = 20, got %f", m.holdings[0].pnlPct)
	}
}

func TestNoPriceFallsBackToCostBasis(t *testing.T) {
	m := New()
	m.width = 120
	m.height = 30
	m.SetHoldings([]store.Holding{
		{Symbol: "AAPL", Quantity: 10, AvgCostBasis: 150.0},
	})

	// No UpdateQuotes — price stays 0
	// Value should fall back to cost basis: 10 * 150 = 1500
	if !approxEqual(m.TotalValue(), 1500.0, 0.01) {
		t.Errorf("expected fallback value 1500, got %f", m.TotalValue())
	}
}
