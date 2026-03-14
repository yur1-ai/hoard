package portfolio

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func testKeyPress(k string) tea.Msg {
	switch k {
	case "tab":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	case "esc":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	case "enter":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	case "shift+tab":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift})
	default:
		if len(k) == 1 {
			return tea.KeyPressMsg(tea.Key{Code: rune(k[0]), Text: k})
		}
		return tea.KeyPressMsg(tea.Key{})
	}
}

func TestFormValidation(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		qty     string
		cost    string
		market  string
		wantErr string
	}{
		{"empty symbol", "", "10", "150", "us_equity", "Symbol is required"},
		{"zero quantity", "AAPL", "0", "150", "us_equity", "Quantity must be a positive number"},
		{"negative quantity", "AAPL", "-5", "150", "us_equity", "Quantity must be a positive number"},
		{"non-numeric quantity", "AAPL", "abc", "150", "us_equity", "Quantity must be a positive number"},
		{"zero cost", "AAPL", "10", "0", "us_equity", "Avg cost must be a positive number"},
		{"non-numeric cost", "AAPL", "10", "abc", "us_equity", "Avg cost must be a positive number"},
		{"invalid market", "AAPL", "10", "150", "forex", "Market must be us_equity or crypto"},
		{"valid us_equity", "AAPL", "10", "150.50", "us_equity", ""},
		{"valid crypto", "BTC", "0.5", "50000", "crypto", ""},
		{"empty market defaults", "AAPL", "10", "150", "", ""},
		{"symbol uppercased", "aapl", "10", "150", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := newAddForm(80)
			f.inputs[fieldSymbol].SetValue(tt.symbol)
			f.inputs[fieldQuantity].SetValue(tt.qty)
			f.inputs[fieldCost].SetValue(tt.cost)
			f.inputs[fieldMarket].SetValue(tt.market)

			result, errMsg := f.validate()
			if tt.wantErr != "" {
				if errMsg != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, errMsg)
				}
				return
			}
			if errMsg != "" {
				t.Fatalf("unexpected error: %s", errMsg)
			}

			// Verify parsed values for valid cases
			if tt.name == "symbol uppercased" && result.Symbol != "AAPL" {
				t.Errorf("expected symbol AAPL, got %s", result.Symbol)
			}
			if tt.name == "empty market defaults" && result.Market != "us_equity" {
				t.Errorf("expected market us_equity, got %s", result.Market)
			}
			if tt.name == "valid crypto" && result.Market != "crypto" {
				t.Errorf("expected market crypto, got %s", result.Market)
			}
		})
	}
}

func TestFormShowHide(t *testing.T) {
	m := New()
	m.SetSize(120, 30)

	if m.IsShowingForm() {
		t.Fatal("form should not show initially")
	}

	m.ShowForm()
	if !m.IsShowingForm() {
		t.Error("expected form to be visible after ShowForm")
	}

	// View should render form content
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when form is shown")
	}

	m.HideForm()
	if m.IsShowingForm() {
		t.Error("expected form hidden after HideForm")
	}
}

func TestFormEscEmitsCancelMsg(t *testing.T) {
	f, _ := newAddForm(80)

	// Simulate Esc key
	_, cmd := f.Update(testKeyPress("esc"))
	if cmd == nil {
		t.Fatal("expected a command from Esc")
	}

	msg := cmd()
	if _, ok := msg.(CancelFormMsg); !ok {
		t.Errorf("expected CancelFormMsg, got %T", msg)
	}
}

func TestFormTabAdvancesField(t *testing.T) {
	f, _ := newAddForm(80)

	if f.focused != fieldSymbol {
		t.Fatalf("expected initial focus on symbol, got %d", f.focused)
	}

	f, _ = f.Update(testKeyPress("tab"))
	if f.focused != fieldQuantity {
		t.Errorf("expected focus on quantity after tab, got %d", f.focused)
	}

	f, _ = f.Update(testKeyPress("tab"))
	if f.focused != fieldCost {
		t.Errorf("expected focus on cost after second tab, got %d", f.focused)
	}
}

func TestFormShiftTabGoesBack(t *testing.T) {
	f, _ := newAddForm(80)
	f, _ = f.Update(testKeyPress("tab"))
	f, _ = f.Update(testKeyPress("tab"))

	if f.focused != fieldCost {
		t.Fatalf("setup: expected focus on cost, got %d", f.focused)
	}

	f, _ = f.Update(testKeyPress("shift+tab"))
	if f.focused != fieldQuantity {
		t.Errorf("expected focus on quantity after shift+tab, got %d", f.focused)
	}
}

func TestFormShiftTabAtFirstFieldStays(t *testing.T) {
	f, _ := newAddForm(80)

	f, _ = f.Update(testKeyPress("shift+tab"))
	if f.focused != fieldSymbol {
		t.Errorf("expected focus to stay on symbol, got %d", f.focused)
	}
}

func TestFormSubmitOnLastFieldEnter(t *testing.T) {
	f, _ := newAddForm(80)
	f.inputs[fieldSymbol].SetValue("AAPL")
	f.inputs[fieldQuantity].SetValue("10")
	f.inputs[fieldCost].SetValue("150")
	f.inputs[fieldMarket].SetValue("us_equity")

	// Move to last field
	f.focused = fieldMarket

	f, cmd := f.Update(testKeyPress("enter"))
	if cmd == nil {
		t.Fatal("expected a command from enter on last field")
	}

	msg := cmd()
	addMsg, ok := msg.(AddHoldingMsg)
	if !ok {
		t.Fatalf("expected AddHoldingMsg, got %T", msg)
	}
	if addMsg.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", addMsg.Symbol)
	}
	if addMsg.Quantity != 10 {
		t.Errorf("expected quantity 10, got %f", addMsg.Quantity)
	}
}

func TestFormSubmitWithValidationError(t *testing.T) {
	f, _ := newAddForm(80)
	// Leave symbol empty
	f.inputs[fieldQuantity].SetValue("10")
	f.inputs[fieldCost].SetValue("150")

	// Move to last field and press enter
	f.focused = fieldMarket
	f, cmd := f.Update(testKeyPress("enter"))

	if cmd != nil {
		t.Error("expected no command when validation fails")
	}
	if f.err == "" {
		t.Error("expected validation error message")
	}
}
