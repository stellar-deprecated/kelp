package kelpdb

/*
	tables
*/
const SqlMarketsTableCreate = "CREATE TABLE IF NOT EXISTS markets (market_id TEXT PRIMARY KEY, exchange_name TEXT NOT NULL, base TEXT NOT NULL, quote TEXT NOT NULL)"
const SqlTradesTableCreate = "CREATE TABLE IF NOT EXISTS trades (market_id TEXT NOT NULL, txid TEXT NOT NULL, date_utc TIMESTAMP WITHOUT TIME ZONE NOT NULL, action TEXT NOT NULL, type TEXT NOT NULL, counter_price DOUBLE PRECISION NOT NULL, base_volume DOUBLE PRECISION NOT NULL, counter_cost DOUBLE PRECISION NOT NULL, fee DOUBLE PRECISION NOT NULL, PRIMARY KEY (market_id, txid))"

/*
	indexes
*/
const SqlTradesIndexCreate = "CREATE INDEX IF NOT EXISTS date ON trades (market_id, date_utc)"
const SqlTradesIndexDrop = "DROP INDEX IF EXISTS date"
const SqlTradesIndexCreate2 = "CREATE INDEX IF NOT EXISTS trades_mdd ON trades (market_id, DATE(date_utc), date_utc)"

/*
	insert statements
*/
// SqlMarketsInsertTemplate inserts into the markets table
const SqlMarketsInsertTemplate = "INSERT INTO markets (market_id, exchange_name, base, quote) VALUES ('%s', '%s', '%s', '%s')"

// SqlTradesInsertTemplate inserts into the trades table
const SqlTradesInsertTemplate = "INSERT INTO trades (market_id, txid, date_utc, action, type, counter_price, base_volume, counter_cost, fee) VALUES ('%s', '%s', '%s', '%s', '%s', %.15f, %.15f, %.15f, %.15f)"

/*
	queries
*/
// SqlQueryMarketsById queries the markets table
const SqlQueryMarketsById = "SELECT market_id, exchange_name, base, quote FROM markets WHERE market_id = $1 LIMIT 1"

// SqlQueryDailyValuesTemplate queries the trades table to get the values for a given day
const SqlQueryDailyValuesTemplate = "SELECT SUM(base_volume) as total_base_volume, SUM(counter_cost) as total_counter_volume FROM trades WHERE market_id IN (%s) AND DATE(date_utc) = $1 and action = $2 group by DATE(date_utc)"
