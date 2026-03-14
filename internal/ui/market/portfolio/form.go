package portfolio

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/yur1-ai/hoard/internal/ui/common"
)

type formField int

const (
	fieldSymbol formField = iota
	fieldQuantity
	fieldCost
	fieldMarket
	fieldCount
)

// AddHoldingMsg is emitted when the user submits the add-position form.
type AddHoldingMsg struct {
	Symbol   string
	Quantity float64
	AvgCost  float64
	Market   string
}

// CancelFormMsg is emitted when the user cancels the form.
type CancelFormMsg struct{}

type addForm struct {
	inputs  [fieldCount]textinput.Model
	focused formField
	err     string
	width   int
}

func newAddForm(width int) (addForm, tea.Cmd) {
	var inputs [fieldCount]textinput.Model

	symbol := textinput.New()
	symbol.Placeholder = "AAPL"
	symbol.Prompt = "Symbol:   "
	symbol.CharLimit = 12
	cmd := symbol.Focus()
	inputs[fieldSymbol] = symbol

	qty := textinput.New()
	qty.Placeholder = "10"
	qty.Prompt = "Quantity: "
	qty.CharLimit = 15
	inputs[fieldQuantity] = qty

	cost := textinput.New()
	cost.Placeholder = "150.00"
	cost.Prompt = "Avg Cost: "
	cost.CharLimit = 15
	inputs[fieldCost] = cost

	mkt := textinput.New()
	mkt.Placeholder = "us_equity"
	mkt.Prompt = "Market:   "
	mkt.CharLimit = 15
	inputs[fieldMarket] = mkt

	return addForm{inputs: inputs, width: width}, cmd
}

func (f addForm) Update(msg tea.Msg) (addForm, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return CancelFormMsg{} }

		case "tab", "enter":
			// On last field enter → validate and submit
			if f.focused == fieldCount-1 {
				result, err := f.validate()
				if err != "" {
					f.err = err
					return f, nil
				}
				return f, func() tea.Msg { return result }
			}
			// Advance to next field
			f.inputs[f.focused].Blur()
			f.focused++
			cmd := f.inputs[f.focused].Focus()
			return f, cmd

		case "shift+tab":
			if f.focused > 0 {
				f.inputs[f.focused].Blur()
				f.focused--
				cmd := f.inputs[f.focused].Focus()
				return f, cmd
			}
			return f, nil
		}
	}

	// Forward to focused input
	var cmd tea.Cmd
	f.inputs[f.focused], cmd = f.inputs[f.focused].Update(msg)
	f.err = "" // clear error on new input
	return f, cmd
}

func (f addForm) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(common.ColorHighlight)
	errStyle := lipgloss.NewStyle().Foreground(common.ColorRed)

	var b strings.Builder
	b.WriteString(titleStyle.Render("  ADD POSITION"))
	b.WriteString("\n\n")

	for i := range fieldCount {
		b.WriteString("  ")
		b.WriteString(f.inputs[i].View())
		b.WriteString("\n")
	}

	if f.err != "" {
		fmt.Fprintf(&b, "\n  %s", errStyle.Render(f.err))
	}

	b.WriteString("\n  ")
	b.WriteString(common.MutedStyle.Render("Tab: next field • Enter: submit • Esc: cancel"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorHighlight).
		Padding(1, 2).
		Width(f.width - 4)

	return box.Render(b.String())
}

func (f addForm) validate() (AddHoldingMsg, string) {
	symbol := strings.TrimSpace(strings.ToUpper(f.inputs[fieldSymbol].Value()))
	if symbol == "" {
		return AddHoldingMsg{}, "Symbol is required"
	}

	qty, err := strconv.ParseFloat(strings.TrimSpace(f.inputs[fieldQuantity].Value()), 64)
	if err != nil || qty <= 0 {
		return AddHoldingMsg{}, "Quantity must be a positive number"
	}

	cost, err := strconv.ParseFloat(strings.TrimSpace(f.inputs[fieldCost].Value()), 64)
	if err != nil || cost <= 0 {
		return AddHoldingMsg{}, "Avg cost must be a positive number"
	}

	mkt := strings.TrimSpace(strings.ToLower(f.inputs[fieldMarket].Value()))
	if mkt == "" {
		mkt = "us_equity"
	}
	validMarkets := map[string]bool{"us_equity": true, "crypto": true}
	if !validMarkets[mkt] {
		return AddHoldingMsg{}, "Market must be us_equity or crypto"
	}

	return AddHoldingMsg{
		Symbol:   symbol,
		Quantity: qty,
		AvgCost:  cost,
		Market:   mkt,
	}, ""
}
