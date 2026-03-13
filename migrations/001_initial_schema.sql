CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('brokerage','retirement','crypto_wallet')),
    currency TEXT NOT NULL DEFAULT 'USD',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS holdings (
    id INTEGER PRIMARY KEY,
    account_id INTEGER REFERENCES accounts(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    quantity REAL NOT NULL,
    avg_cost_basis REAL NOT NULL,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY,
    account_id INTEGER REFERENCES accounts(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    type TEXT NOT NULL CHECK(type IN ('buy','sell','dividend','transfer')),
    quantity REAL NOT NULL,
    price REAL NOT NULL,
    fee REAL DEFAULT 0,
    date DATETIME NOT NULL,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS watchlists (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS watchlist_items (
    watchlist_id INTEGER REFERENCES watchlists(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    market TEXT NOT NULL CHECK(market IN ('us_equity','crypto')),
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (watchlist_id, symbol)
);

CREATE TABLE IF NOT EXISTS market_cache (
    symbol TEXT PRIMARY KEY,
    market TEXT NOT NULL,
    price REAL,
    change REAL,
    change_pct REAL,
    volume REAL,
    high_24h REAL,
    low_24h REAL,
    last_updated DATETIME
);

CREATE TABLE IF NOT EXISTS currency_rates (
    from_currency TEXT NOT NULL,
    to_currency TEXT NOT NULL,
    rate REAL NOT NULL,
    fetched_at DATETIME NOT NULL,
    PRIMARY KEY (from_currency, to_currency)
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'todo' CHECK(status IN ('todo','done')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

CREATE TABLE IF NOT EXISTS standup_entries (
    id INTEGER PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    yesterday TEXT,
    today TEXT,
    blockers TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS calendar_cache (
    event_id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    calendar_name TEXT,
    last_synced DATETIME
);
