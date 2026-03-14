package store

import (
	"testing"
	"time"
)

func TestCreateAccountAndHolding(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	accountID, err := CreateAccount(db, "Brokerage", "brokerage", "USD")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	if accountID == 0 {
		t.Error("expected non-zero account ID")
	}

	h := Holding{
		AccountID:    accountID,
		Symbol:       "AAPL",
		Market:       "us_equity",
		Quantity:     10,
		AvgCostBasis: 150.00,
		Notes:        "initial buy",
	}
	holdingID, err := CreateHolding(db, h)
	if err != nil {
		t.Fatalf("create holding: %v", err)
	}
	if holdingID == 0 {
		t.Error("expected non-zero holding ID")
	}

	holdings, err := ListHoldings(db, accountID)
	if err != nil {
		t.Fatalf("list holdings: %v", err)
	}
	if len(holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", holdings[0].Symbol)
	}
	if holdings[0].Quantity != 10 {
		t.Errorf("expected qty 10, got %f", holdings[0].Quantity)
	}
}

func TestListAccounts(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	CreateAccount(db, "Brokerage", "brokerage", "USD")
	CreateAccount(db, "Crypto", "crypto_wallet", "USD")

	accounts, err := ListAccounts(db)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}
}

func TestListHoldingsAllAccounts(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	id1, _ := CreateAccount(db, "A1", "brokerage", "USD")
	id2, _ := CreateAccount(db, "A2", "brokerage", "USD")

	CreateHolding(db, Holding{AccountID: id1, Symbol: "AAPL", Market: "us_equity", Quantity: 10, AvgCostBasis: 150})
	CreateHolding(db, Holding{AccountID: id2, Symbol: "GOOG", Market: "us_equity", Quantity: 5, AvgCostBasis: 100})

	// accountID=0 means all accounts
	holdings, err := ListHoldings(db, 0)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(holdings) != 2 {
		t.Errorf("expected 2 holdings across accounts, got %d", len(holdings))
	}
}

func TestAddBuyTransaction(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	acctID, _ := CreateAccount(db, "Main", "brokerage", "USD")

	// First buy — should auto-create holding
	tx1 := Transaction{
		AccountID: acctID,
		Symbol:    "AAPL",
		Market:    "us_equity",
		Type:      "buy",
		Quantity:  10,
		Price:     100.00,
		Fee:       0,
		Date:      time.Now(),
	}
	txID, err := AddTransaction(db, tx1)
	if err != nil {
		t.Fatalf("add buy: %v", err)
	}
	if txID == 0 {
		t.Error("expected non-zero tx ID")
	}

	holdings, _ := ListHoldings(db, acctID)
	if len(holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(holdings))
	}
	if holdings[0].Quantity != 10 {
		t.Errorf("expected qty 10, got %f", holdings[0].Quantity)
	}
	if holdings[0].AvgCostBasis != 100.00 {
		t.Errorf("expected avg cost 100, got %f", holdings[0].AvgCostBasis)
	}

	// Second buy — should update avg cost basis
	tx2 := Transaction{
		AccountID: acctID,
		Symbol:    "AAPL",
		Market:    "us_equity",
		Type:      "buy",
		Quantity:  10,
		Price:     200.00,
		Date:      time.Now(),
	}
	AddTransaction(db, tx2)

	holdings, _ = ListHoldings(db, acctID)
	if holdings[0].Quantity != 20 {
		t.Errorf("expected qty 20, got %f", holdings[0].Quantity)
	}
	// avg = ((10*100) + (10*200)) / 20 = 3000/20 = 150
	if holdings[0].AvgCostBasis != 150.00 {
		t.Errorf("expected avg cost 150, got %f", holdings[0].AvgCostBasis)
	}
}

func TestSellTransaction(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	acctID, _ := CreateAccount(db, "Main", "brokerage", "USD")

	// Buy first
	AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "buy", Quantity: 20, Price: 100.00, Date: time.Now(),
	})

	// Sell half
	_, err = AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "sell", Quantity: 5, Price: 150.00, Date: time.Now(),
	})
	if err != nil {
		t.Fatalf("sell: %v", err)
	}

	holdings, _ := ListHoldings(db, acctID)
	if holdings[0].Quantity != 15 {
		t.Errorf("expected qty 15 after sell, got %f", holdings[0].Quantity)
	}
	// Avg cost basis unchanged on sell
	if holdings[0].AvgCostBasis != 100.00 {
		t.Errorf("expected avg cost unchanged at 100, got %f", holdings[0].AvgCostBasis)
	}
}

func TestSellMoreThanOwned(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	acctID, _ := CreateAccount(db, "Main", "brokerage", "USD")

	AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "buy", Quantity: 10, Price: 100.00, Date: time.Now(),
	})

	// Try to sell more than owned
	_, err = AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "sell", Quantity: 20, Price: 150.00, Date: time.Now(),
	})
	if err == nil {
		t.Error("expected error when selling more than owned")
	}
}

func TestSellWithNoHolding(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	acctID, _ := CreateAccount(db, "Main", "brokerage", "USD")

	_, err = AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "sell", Quantity: 5, Price: 150.00, Date: time.Now(),
	})
	if err == nil {
		t.Error("expected error when selling with no holding")
	}
}

func TestDeleteHolding(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	acctID, _ := CreateAccount(db, "Main", "brokerage", "USD")
	hID, _ := CreateHolding(db, Holding{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Quantity: 10, AvgCostBasis: 100,
	})

	if err := DeleteHolding(db, hID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	holdings, _ := ListHoldings(db, acctID)
	if len(holdings) != 0 {
		t.Errorf("expected 0 holdings after delete, got %d", len(holdings))
	}
}

func TestListTransactions(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	acctID, _ := CreateAccount(db, "Main", "brokerage", "USD")

	AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "buy", Quantity: 10, Price: 100, Date: time.Now(),
	})
	AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "buy", Quantity: 5, Price: 110, Date: time.Now(),
	})
	AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "GOOG", Market: "us_equity",
		Type: "buy", Quantity: 3, Price: 200, Date: time.Now(),
	})

	txs, err := ListTransactions(db, "AAPL")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 AAPL txs, got %d", len(txs))
	}
}

func TestAllEquityAndCryptoSymbols(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	acctID, _ := CreateAccount(db, "Main", "brokerage", "USD")
	CreateHolding(db, Holding{AccountID: acctID, Symbol: "AAPL", Market: "us_equity", Quantity: 10, AvgCostBasis: 100})
	CreateHolding(db, Holding{AccountID: acctID, Symbol: "GOOG", Market: "us_equity", Quantity: 5, AvgCostBasis: 200})
	CreateHolding(db, Holding{AccountID: acctID, Symbol: "BTC", Market: "crypto", Quantity: 1, AvgCostBasis: 60000})

	equities, err := AllEquitySymbols(db)
	if err != nil {
		t.Fatalf("equity symbols: %v", err)
	}
	if len(equities) != 2 {
		t.Errorf("expected 2 equity symbols, got %d", len(equities))
	}

	cryptos, err := AllCryptoSymbols(db)
	if err != nil {
		t.Fatalf("crypto symbols: %v", err)
	}
	if len(cryptos) != 1 {
		t.Errorf("expected 1 crypto symbol, got %d", len(cryptos))
	}
}

func TestDividendTransaction(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	acctID, _ := CreateAccount(db, "Main", "brokerage", "USD")

	// Buy first to have a holding
	AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "buy", Quantity: 10, Price: 100, Date: time.Now(),
	})

	// Dividend doesn't change holding quantity or avg cost
	_, err = AddTransaction(db, Transaction{
		AccountID: acctID, Symbol: "AAPL", Market: "us_equity",
		Type: "dividend", Quantity: 0, Price: 5.00, Date: time.Now(),
	})
	if err != nil {
		t.Fatalf("dividend: %v", err)
	}

	holdings, _ := ListHoldings(db, acctID)
	if holdings[0].Quantity != 10 {
		t.Errorf("dividend should not change qty, got %f", holdings[0].Quantity)
	}
}
