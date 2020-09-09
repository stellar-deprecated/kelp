package kelpdb

/*
	tables
*/
const SqlMarketsTableCreate = "CREATE TABLE IF NOT EXISTS markets (market_id TEXT PRIMARY KEY, exchange_name TEXT NOT NULL, base TEXT NOT NULL, quote TEXT NOT NULL)"
const SqlTradesTableCreate = "CREATE TABLE IF NOT EXISTS trades (market_id TEXT NOT NULL, txid TEXT NOT NULL, date_utc TIMESTAMP WITHOUT TIME ZONE NOT NULL, action TEXT NOT NULL, type TEXT NOT NULL, counter_price DOUBLE PRECISION NOT NULL, base_volume DOUBLE PRECISION NOT NULL, counter_cost DOUBLE PRECISION NOT NULL, fee DOUBLE PRECISION NOT NULL, PRIMARY KEY (market_id, txid))"
const SqlTradesTableAlter1 = "ALTER TABLE trades ADD COLUMN account_id TEXT"
const SqlStrategyMirrorTradeTriggersTableCreate = "CREATE TABLE IF NOT EXISTS strategy_mirror_trade_triggers (market_id TEXT NOT NULL, txid TEXT NOT NULL, backing_market_id TEXT NOT NULL, backing_order_id TEXT NOT NULL, PRIMARY KEY (market_id, txid))"
const SqlTradesTableAlter2 = "ALTER TABLE trades ADD COLUMN order_id TEXT"

/*
	indexes
*/
const SqlTradesIndexCreate = "CREATE INDEX IF NOT EXISTS date ON trades (market_id, date_utc)"
const SqlTradesIndexDrop = "DROP INDEX IF EXISTS date"
const SqlTradesIndexCreate2 = "CREATE INDEX IF NOT EXISTS trades_mdd ON trades (market_id, DATE(date_utc), date_utc)"

// We don't include account_id in the primary key of the trades table because the account_id will initially be null until we clean that up (later)
// For now we add it as a unique index on which we will later base the primary key. This does not provide us with any immediate benefit because the PK is a subset
// of this unique index and we don't use this index for queries yet (we will later)
const SqlTradesIndexCreate3 = "CREATE UNIQUE INDEX IF NOT EXISTS trades_amt ON trades (account_id, market_id, txid)"

/*
	insert statements
*/
// SqlMarketsInsertTemplate inserts into the markets table
const SqlMarketsInsertTemplate = "INSERT INTO markets (market_id, exchange_name, base, quote) VALUES ('%s', '%s', '%s', '%s')"

// SqlTradesInsertTemplate inserts into the trades table
const SqlTradesInsertTemplate = "INSERT INTO trades (market_id, txid, date_utc, action, type, counter_price, base_volume, counter_cost, fee, account_id, order_id) VALUES ('%s', '%s', '%s', '%s', '%s', %.15f, %.15f, %.15f, %.15f, '%s', '%s')"

// SqlStrategyMirrorTradeTriggersInsertTemplate inserts into the strategy_mirror_trade_triggers table
const SqlStrategyMirrorTradeTriggersInsertTemplate = "INSERT INTO strategy_mirror_trade_triggers (market_id, txid, backing_market_id, backing_order_id) VALUES ('%s', '%s', '%s', '%s')"

/*
	queries
*/
// SqlQueryMarketsById queries the markets table
const SqlQueryMarketsById = "SELECT market_id, exchange_name, base, quote FROM markets WHERE market_id = $1 LIMIT 1"
