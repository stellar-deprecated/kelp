package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/support/database"
)

func TestTradeUpgradeScripts(t *testing.T) {
	// run the PreTest and defer running the postTest
	db, dbname := database.PreTest(t)
	defer database.PostTestWithDbClose(db, dbname)

	// run the upgrade scripts
	codeVersionString := "TestTradeUpgradeScripts"
	e := database.RunUpgradeScripts(db, upgradeScripts, codeVersionString)
	if e != nil {
		panic(e)
	}

	// assert current state of the database
	assert.Equal(t, 4, database.GetNumTablesInDb(db))
	assert.True(t, database.CheckTableExists(db, "db_version"))
	assert.True(t, database.CheckTableExists(db, "markets"))
	assert.True(t, database.CheckTableExists(db, "trades"))
	assert.True(t, database.CheckTableExists(db, "strategy_mirror_trade_triggers"))

	// check schema of db_version table
	var columns []database.TableColumn
	columns = database.GetTableSchema(db, "db_version")
	assert.Equal(t, 5, len(columns), fmt.Sprintf("%v", columns))
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "version",
		OrdinalPosition:        1,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "integer",
		CharacterMaximumLength: nil,
	}, &columns[0])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "date_completed_utc",
		OrdinalPosition:        2,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "timestamp without time zone",
		CharacterMaximumLength: nil,
	}, &columns[1])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "num_scripts",
		OrdinalPosition:        3,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "integer",
		CharacterMaximumLength: nil,
	}, &columns[2])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "time_elapsed_millis",
		OrdinalPosition:        4,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "bigint",
		CharacterMaximumLength: nil,
	}, &columns[3])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "code_version_string",
		OrdinalPosition:        5,
		ColumnDefault:          nil,
		IsNullable:             "YES",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[4])
	// check indexes of db_version table
	indexes := database.GetTableIndexes(db, "db_version")
	assert.Equal(t, 1, len(indexes))
	database.AssertIndex(t, "db_version", "db_version_pkey", "CREATE UNIQUE INDEX db_version_pkey ON public.db_version USING btree (version)", indexes)

	// check schema of markets table
	columns = database.GetTableSchema(db, "markets")
	assert.Equal(t, 4, len(columns), fmt.Sprintf("%v", columns))
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "market_id",
		OrdinalPosition:        1,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[0])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "exchange_name",
		OrdinalPosition:        2,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[1])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "base",
		OrdinalPosition:        3,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[2])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "quote",
		OrdinalPosition:        4,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[3])
	// check indexes of markets table
	indexes = database.GetTableIndexes(db, "markets")
	assert.Equal(t, 1, len(indexes))
	database.AssertIndex(t, "markets", "markets_pkey", "CREATE UNIQUE INDEX markets_pkey ON public.markets USING btree (market_id)", indexes)

	// check schema of trades table
	columns = database.GetTableSchema(db, "trades")
	assert.Equal(t, 11, len(columns), fmt.Sprintf("%v", columns))
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "market_id",
		OrdinalPosition:        1,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[0])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "txid",
		OrdinalPosition:        2,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[1])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "date_utc",
		OrdinalPosition:        3,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "timestamp without time zone",
		CharacterMaximumLength: nil,
	}, &columns[2])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "action",
		OrdinalPosition:        4,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[3])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "type",
		OrdinalPosition:        5,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[4])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "counter_price",
		OrdinalPosition:        6,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "double precision",
		CharacterMaximumLength: nil,
	}, &columns[5])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "base_volume",
		OrdinalPosition:        7,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "double precision",
		CharacterMaximumLength: nil,
	}, &columns[6])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "counter_cost",
		OrdinalPosition:        8,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "double precision",
		CharacterMaximumLength: nil,
	}, &columns[7])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "fee",
		OrdinalPosition:        9,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "double precision",
		CharacterMaximumLength: nil,
	}, &columns[8])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "account_id",
		OrdinalPosition:        10,
		ColumnDefault:          nil,
		IsNullable:             "YES",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[9])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "order_id",
		OrdinalPosition:        11,
		ColumnDefault:          nil,
		IsNullable:             "YES",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[10])
	// check indexes of trades table
	indexes = database.GetTableIndexes(db, "trades")
	assert.Equal(t, 3, len(indexes))
	database.AssertIndex(t, "trades", "trades_pkey", "CREATE UNIQUE INDEX trades_pkey ON public.trades USING btree (market_id, txid)", indexes)
	database.AssertIndex(t, "trades", "trades_mdd", "CREATE INDEX trades_mdd ON public.trades USING btree (market_id, date(date_utc), date_utc)", indexes)
	database.AssertIndex(t, "trades", "trades_amt", "CREATE UNIQUE INDEX trades_amt ON public.trades USING btree (account_id, market_id, txid)", indexes)

	// check schema of strategy_mirror_trade_triggers table
	columns = database.GetTableSchema(db, "strategy_mirror_trade_triggers")
	assert.Equal(t, 4, len(columns), fmt.Sprintf("%v", columns))
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "market_id",
		OrdinalPosition:        1,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[0])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "txid",
		OrdinalPosition:        2,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[1])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "backing_market_id",
		OrdinalPosition:        3,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[2])
	database.AssertTableColumnsEqual(t, &database.TableColumn{
		ColumnName:             "backing_order_id",
		OrdinalPosition:        4,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[3])
	// check indexes of strategy_mirror_trade_triggers table
	indexes = database.GetTableIndexes(db, "strategy_mirror_trade_triggers")
	assert.Equal(t, 1, len(indexes))
	database.AssertIndex(t, "strategy_mirror_trade_triggers", "strategy_mirror_trade_triggers_pkey", "CREATE UNIQUE INDEX strategy_mirror_trade_triggers_pkey ON public.strategy_mirror_trade_triggers USING btree (market_id, txid)", indexes)

	// check entries of db_version table
	var allRows [][]interface{}
	allRows = database.QueryAllRows(db, "db_version")
	assert.Equal(t, 6, len(allRows))
	// first three code_version_string is nil becuase the field was not supported at the time when the upgrade script was run, and only in version 4 of
	// the database do we add the field. See upgradeScripts and RunUpgradeScripts() for more details
	database.ValidateDBVersionRow(t, allRows[0], 1, time.Now(), 1, 50, nil)
	database.ValidateDBVersionRow(t, allRows[1], 2, time.Now(), 3, 150, nil)
	database.ValidateDBVersionRow(t, allRows[2], 3, time.Now(), 2, 100, nil)
	database.ValidateDBVersionRow(t, allRows[3], 4, time.Now(), 1, 50, &codeVersionString)
	database.ValidateDBVersionRow(t, allRows[4], 5, time.Now(), 2, 100, &codeVersionString)
	database.ValidateDBVersionRow(t, allRows[5], 6, time.Now(), 2, 100, &codeVersionString)

	// check entries of markets table
	allRows = database.QueryAllRows(db, "markets")
	assert.Equal(t, 0, len(allRows))

	// check entries of trades table
	allRows = database.QueryAllRows(db, "trades")
	assert.Equal(t, 0, len(allRows))

	// check entries of strategy_mirror_trade_triggers table
	allRows = database.QueryAllRows(db, "strategy_mirror_trade_triggers")
	assert.Equal(t, 0, len(allRows))
}
