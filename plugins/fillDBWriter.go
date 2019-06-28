package plugins

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

const dateFormatString = "2006/01/02"
const sqlDbCreate = "CREATE TABLE IF NOT EXISTS trades (txid TEXT PRIMARY KEY, date_utc VARCHAR(10), timestamp_millis INTEGER, base TEXT, quote TEXT, action TEXT, type TEXT, counter_price REAL, base_volume REAL, counter_cost REAL, fee REAL)"
const sqlInsert = "INSERT INTO trades (txid, date_utc, timestamp_millis, base, quote, action, type, counter_price, base_volume, counter_cost, fee) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

var sqlIndexes = []string{
	"CREATE INDEX IF NOT EXISTS date ON trades (date_utc, base, quote)",
}

// FillDBWriter is a FillHandler that writes fills to a SQL database
type FillDBWriter struct {
	db *sql.DB
}

var _ api.FillHandler = &FillDBWriter{}

// MakeFillDBWriter is a factory method
func MakeFillDBWriter(sqlDbPath string) (api.FillHandler, error) {
	db, e := sql.Open("sqlite3", sqlDbPath)
	if e != nil {
		return nil, fmt.Errorf("could not open sqlite3 database: %s", e)
	}

	statement, e := db.Prepare(sqlDbCreate)
	if e != nil {
		return nil, fmt.Errorf("could not prepare sql statement: %s", e)
	}
	_, e = statement.Exec()
	if e != nil {
		return nil, fmt.Errorf("could not execute sql create table statement: %s", e)
	}

	for i, sqlIndexCreate := range sqlIndexes {
		statement, e = db.Prepare(sqlIndexCreate)
		if e != nil {
			return nil, fmt.Errorf("could not prepare sql statement to create index (i=%d): %s", i, e)
		}
		_, e = statement.Exec()
		if e != nil {
			return nil, fmt.Errorf("could not execute sql statement to create index (i=%d): %s", i, e)
		}
	}

	log.Printf("making FillDBWriter with db path: %s\n", sqlDbPath)
	return &FillDBWriter{
		db: db,
	}, nil
}

// HandleFill impl.
func (f *FillDBWriter) HandleFill(trade model.Trade) error {
	statement, e := f.db.Prepare(sqlInsert)
	if e != nil {
		return fmt.Errorf("could not prepare sql insert values statement: %s", e)
	}

	txid := utils.CheckedString(trade.TransactionID)
	timeSeconds := trade.Timestamp.AsInt64() / 1000
	date := time.Unix(timeSeconds, 0).UTC()
	dateString := date.Format(dateFormatString)

	_, e = statement.Exec(
		txid,
		dateString,
		utils.CheckedString(trade.Timestamp),
		string(trade.Pair.Base),
		string(trade.Pair.Quote),
		trade.OrderAction.String(),
		trade.OrderType.String(),
		f.checkedFloat(trade.Price),
		f.checkedFloat(trade.Volume),
		f.checkedFloat(trade.Cost),
		f.checkedFloat(trade.Fee),
	)
	if e != nil {
		return fmt.Errorf("could not execute sql insert values statement: %s", e)
	}
	log.Printf("wrote trade (txid=%s) to db\n", txid)
	return nil
}

func (f *FillDBWriter) checkedFloat(n *model.Number) interface{} {
	if n == nil {
		return nil
	}
	return n.AsFloat()
}
