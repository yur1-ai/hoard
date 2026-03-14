package portfolio

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/yur1-ai/hoard/internal/service/market"
	"github.com/yur1-ai/hoard/internal/store"
	"github.com/yur1-ai/hoard/internal/ui/common"
)

// holdingRow merges a DB holding with live price data.
type holdingRow struct {
	holding  store.Holding
	price    float64
	change   float64
	changePct float64
	pnl      float64
	pnlPct   float64
	value    float64
	allocPct float64
}

// Model manages the portfolio view.
type Model struct {
	width, height int
	holdings      []holdingRow
	totalValue    float64
	dayChange     float64
	dayChangePct  float64
	table         table.Model
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// SetHoldings replaces the current holdings list and recalculates derived values.
func (m *Model) SetHoldings(holdings []store.Holding) {
	m.holdings = make([]holdingRow, len(holdings))
	for i, h := range holdings {
		m.holdings[i] = holdingRow{holding: h}
	}
	m.recalculate()
	m.rebuildTable()
}

// UpdateQuotes updates prices for matching symbols and recalculates P&L.
func (m *Model) UpdateQuotes(quotes []market.Quote) {
	priceMap := make(map[string]market.Quote, len(quotes))
	for _, q := range quotes {
		priceMap[q.Symbol] = q
	}

	for i := range m.holdings {
		if q, ok := priceMap[m.holdings[i].holding.Symbol]; ok {
			m.holdings[i].price = q.Price
			m.holdings[i].change = q.Change
			m.holdings[i].changePct = q.ChangePct
		}
	}
	m.recalculate()
	m.rebuildTable()
}

// TotalValue returns the total portfolio market value.
func (m Model) TotalValue() float64 { return m.totalValue }

// DayChange returns the total dollar change today.
func (m Model) DayChange() float64 { return m.dayChange }

// DayChangePct returns the portfolio's total day change percentage.
func (m Model) DayChangePct() float64 { return m.dayChangePct }

func (m *Model) recalculate() {
	m.totalValue = 0
	m.dayChange = 0
	totalCost := 0.0

	for i := range m.holdings {
		h := &m.holdings[i]
		qty := h.holding.Quantity
		if h.price > 0 {
			h.value = qty * h.price
			h.pnl = qty * (h.price - h.holding.AvgCostBasis)
			if h.holding.AvgCostBasis > 0 {
				h.pnlPct = ((h.price - h.holding.AvgCostBasis) / h.holding.AvgCostBasis) * 100
			}
		} else {
			// No live price — use cost basis as value estimate
			h.value = qty * h.holding.AvgCostBasis
		}
		m.totalValue += h.value
		m.dayChange += qty * h.change
		totalCost += qty * h.holding.AvgCostBasis
	}

	// Update allocation percentages
	for i := range m.holdings {
		if m.totalValue > 0 {
			m.holdings[i].allocPct = (m.holdings[i].value / m.totalValue) * 100
		}
	}

	// Total day change percent
	if totalCost > 0 {
		m.dayChangePct = (m.dayChange / totalCost) * 100
	}
}

func (m *Model) rebuildTable() {
	columns := []table.Column{
		{Title: "Symbol", Width: 8},
		{Title: "Shares", Width: 8},
		{Title: "Avg Cost", Width: 10},
		{Title: "Price", Width: 10},
		{Title: "Day Chg", Width: 10},
		{Title: "P&L", Width: 12},
		{Title: "Value", Width: 12},
		{Title: "Alloc", Width: 7},
	}

	rows := make([]table.Row, len(m.holdings))
	for i, h := range m.holdings {
		priceStr := "—"
		dayStr := "—"
		pnlStr := "—"
		valueStr := "—"
		if h.price > 0 {
			priceStr = fmt.Sprintf("$%.2f", h.price)
			dayStr = common.FormatChange(h.changePct)
			pnlStr = common.FormatMoney(h.pnl)
			valueStr = fmt.Sprintf("$%.2f", h.value)
		}
		rows[i] = table.Row{
			h.holding.Symbol,
			fmt.Sprintf("%.2f", h.holding.Quantity),
			fmt.Sprintf("$%.2f", h.holding.AvgCostBasis),
			priceStr,
			dayStr,
			pnlStr,
			valueStr,
			fmt.Sprintf("%.1f%%", h.allocPct),
		}
	}

	height := m.height - 4 // summary + header + padding
	if height < 3 {
		height = 3
	}
	m.table = common.NewStyledTable(columns, rows, height)
}

func (m Model) View() string {
	if len(m.holdings) == 0 {
		content := lipgloss.NewStyle().
			Foreground(common.ColorMuted).
			Render("No positions yet — press [a] to add a holding")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// Summary line
	summary := fmt.Sprintf("  Total: $%.2f  |  Day: %s (%s)  |  %d positions",
		m.totalValue,
		common.FormatMoney(m.dayChange),
		common.FormatPercent(m.dayChangePct),
		len(m.holdings),
	)
	summaryLine := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorWhite).
		Render(summary)

	return lipgloss.JoinVertical(lipgloss.Left, summaryLine, "", m.table.View())
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.rebuildTable()
}
