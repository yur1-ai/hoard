package market

import "context"

// Quote holds price data for a single symbol.
type Quote struct {
	Symbol    string
	Price     float64
	Change    float64
	ChangePct float64
	Volume    float64
	High      float64
	Low       float64
}

// SymbolMatch represents a search result.
type SymbolMatch struct {
	Symbol      string
	Description string
	Type        string // "Common Stock", "ETF", "Crypto"
}

// StockProvider fetches equity market data.
type StockProvider interface {
	GetQuote(ctx context.Context, symbol string) (*Quote, error)
	GetQuotes(ctx context.Context, symbols []string) ([]Quote, error)
	SearchSymbol(ctx context.Context, query string) ([]SymbolMatch, error)
}

// CryptoProvider fetches cryptocurrency market data.
type CryptoProvider interface {
	GetQuote(ctx context.Context, symbol string) (*Quote, error)
	GetQuotes(ctx context.Context, symbols []string) ([]Quote, error)
}
