package market

import "strings"

// topCoinIDs maps common ticker symbols to CoinGecko IDs.
var topCoinIDs = map[string]string{
	"BTC":   "bitcoin",
	"ETH":   "ethereum",
	"BNB":   "binancecoin",
	"SOL":   "solana",
	"XRP":   "ripple",
	"ADA":   "cardano",
	"DOGE":  "dogecoin",
	"AVAX":  "avalanche-2",
	"DOT":   "polkadot",
	"LINK":  "chainlink",
	"MATIC": "matic-network",
	"SHIB":  "shiba-inu",
	"UNI":   "uniswap",
	"LTC":   "litecoin",
	"ATOM":  "cosmos",
	"NEAR":  "near",
	"APT":   "aptos",
	"ARB":   "arbitrum",
	"OP":    "optimism",
	"FIL":   "filecoin",
	"AAVE":  "aave",
	"MKR":   "maker",
	"GRT":   "the-graph",
	"IMX":   "immutable-x",
	"SAND":  "the-sandbox",
	"MANA":  "decentraland",
	"AXS":   "axie-infinity",
	"CRV":   "curve-dao-token",
	"COMP":  "compound-governance-token",
	"ALGO":  "algorand",
	"SUI":   "sui",
	"SEI":   "sei-network",
	"TIA":   "celestia",
	"PEPE":  "pepe",
	"WIF":   "dogwifcoin",
}

// resolveCoinID maps a ticker symbol (e.g., "BTC" or "BTC-USD") to a CoinGecko ID.
func resolveCoinID(symbol string) (string, bool) {
	ticker := strings.TrimSuffix(strings.ToUpper(symbol), "-USD")
	id, ok := topCoinIDs[ticker]
	return id, ok
}
