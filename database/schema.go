package database

import (
	"database/sql"
	"fmt"
	"log"
)

/*
	tables
*/
const sqlDbVersionTableCreate = "CREATE TABLE IF NOT EXISTS db_version (version INTEGER NOT NULL, date_completed_utc TIMESTAMP WITHOUT TIME ZONE NOT NULL, num_scripts INTEGER NOT NULL, time_elapsed_millis BIGINT NOT NULL, PRIMARY KEY (version))"
const sqlMarketsTableCreate = "CREATE TABLE IF NOT EXISTS markets (market_id TEXT PRIMARY KEY, exchange_name TEXT NOT NULL, base TEXT NOT NULL, quote TEXT NOT NULL)"
const sqlTradesTableCreate = "CREATE TABLE IF NOT EXISTS trades (market_id TEXT NOT NULL, txid TEXT NOT NULL, date_utc TIMESTAMP WITHOUT TIME ZONE NOT NULL, action TEXT NOT NULL, type TEXT NOT NULL, counter_price DOUBLE PRECISION NOT NULL, base_volume DOUBLE PRECISION NOT NULL, counter_cost DOUBLE PRECISION NOT NULL, fee DOUBLE PRECISION NOT NULL, PRIMARY KEY (market_id, txid))"

/*
	indexes
*/
const sqlTradesIndexCreate = "CREATE INDEX IF NOT EXISTS date ON trades (market_id, date_utc)"
const sqlTradesIndexDrop = "DROP INDEX IF EXISTS date"
const sqlTradesIndexCreate2 = "CREATE INDEX IF NOT EXISTS trades_mdd ON trades (market_id, DATE(date_utc), date_utc)"

/*
	insert statements
*/
// sqlDbVersionTableInsertTemplate inserts into the db_version table
const sqlDbVersionTableInsertTemplate = "INSERT INTO db_version (version, date_completed_utc, num_scripts, time_elapsed_millis) VALUES (%d, '%s', %d, %d)"

// SqlMarketsInsertTemplate inserts into the markets table
const SqlMarketsInsertTemplate = "INSERT INTO markets (market_id, exchange_name, base, quote) VALUES ('%s', '%s', '%s', '%s')"

// SqlTradesInsertTemplate inserts into the trades table
const SqlTradesInsertTemplate = "INSERT INTO trades (market_id, txid, date_utc, action, type, counter_price, base_volume, counter_cost, fee) VALUES ('%s', '%s', '%s', '%s', '%s', %.15f, %.15f, %.15f, %.15f)"

/*
	queries
*/
// SqlQueryMarketsById queries the markets table
const SqlQueryMarketsById = "SELECT market_id, exchange_name, base, quote FROM markets WHERE market_id = $1 LIMIT 1"

// sqlQueryDbVersion queries the db_version table
const sqlQueryDbVersion = "SELECT version FROM db_version ORDER BY version desc LIMIT 1"

// SqlQueryDailyValues queries the trades table to get the values for a given day
const SqlQueryDailyValues = "SELECT SUM(base_volume) as total_base_volume, SUM(counter_cost) as total_counter_volume FROM trades WHERE market_id = $1 AND DATE(date_utc) = $2 and action = $3 group by DATE(date_utc)"

/*
	query helper functions
*/
// QueryDbVersion queries for the version of the database
func QueryDbVersion(db *sql.DB) (uint32, error) {
	rows, e := db.Query(sqlQueryDbVersion)
	if e != nil {
		return 0, fmt.Errorf("could not execute sql select query (%s): %s", sqlQueryDbVersion, e)
	}
	defer rows.Close()

	for rows.Next() {
		var dbVersion uint32
		e = rows.Scan(&dbVersion)
		if e != nil {
			return 0, fmt.Errorf("could not scan row to get the db version: %s", e)
		}

		log.Printf("fetched dbVersion from db: %d", dbVersion)
		return dbVersion, nil
	}

	return 0, nil
}
