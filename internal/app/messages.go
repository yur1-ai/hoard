package app

import (
	"time"

	"github.com/yur1-ai/hoard/internal/service/market"
)

// TickMsg fires on each refresh interval.
type TickMsg time.Time

// QuotesMsg carries fetched stock quotes into Update.
type QuotesMsg struct {
	Quotes []market.Quote
}

// CryptoQuotesMsg carries fetched crypto quotes into Update.
type CryptoQuotesMsg struct {
	Quotes []market.Quote
}

// CurrencyRatesMsg carries fetched currency rates into Update.
type CurrencyRatesMsg struct {
	Rates map[string]float64
}

// ErrMsg carries a non-fatal error to display in the status bar.
type ErrMsg struct {
	Err     error
	Context string
}

func (e ErrMsg) Error() string {
	if e.Err == nil {
		return e.Context
	}
	return e.Err.Error()
}
