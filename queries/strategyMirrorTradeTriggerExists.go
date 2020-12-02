package queries

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/utils"
)

// sqlQueryStrategyMirrorTradeTriggerExists queries the strategy_mirror_trade_triggers table by market_id and txid (primary key) to see if the row exists
const sqlQueryStrategyMirrorTradeTriggerExists = "SELECT * FROM strategy_mirror_trade_triggers WHERE market_id = $1 AND txid = $2"

// StrategyMirrorTradeTriggerExists is a query that fetches the row by primary key
type StrategyMirrorTradeTriggerExists struct {
	db       *sql.DB
	sqlQuery string
	marketID string
}

var _ api.Query = &StrategyMirrorTradeTriggerExists{}

// MakeStrategyMirrorTradeTriggerExists makes the StrategyMirrorTradeTriggerExists query
func MakeStrategyMirrorTradeTriggerExists(db *sql.DB, marketID string) (*StrategyMirrorTradeTriggerExists, error) {
	if db == nil {
		utils.PrintErrorHintf("the provided POSTGRES_DB config in the trader.cfg file should be non-nil")
		return nil, fmt.Errorf("the provided db should be non-nil")
	}

	return &StrategyMirrorTradeTriggerExists{
		db:       db,
		sqlQuery: sqlQueryStrategyMirrorTradeTriggerExists,
		marketID: marketID,
	}, nil
}

// Name impl.
func (q *StrategyMirrorTradeTriggerExists) Name() string {
	return "StrategyMirrorTradeTriggerExists"
}

// QueryRow impl.
func (q *StrategyMirrorTradeTriggerExists) QueryRow(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 args (txid string), but got args %v", args)
	} else if _, ok := args[0].(string); !ok {
		return nil, fmt.Errorf("input arg[0] needs to be of type 'string', but was of type '%T'", args[0])
	}

	row := q.db.QueryRow(q.sqlQuery, q.marketID, args[0])
	var marketID, txID, backingMarketID, backingOrderID string
	e := row.Scan(&marketID, &txID, &backingMarketID, &backingOrderID)
	if e != nil {
		if strings.Contains(e.Error(), "no rows in result set") {
			return false, nil
		}
		return nil, fmt.Errorf("could not read data from StrategyMirrorTradeTriggerExists query: %s", e)
	}
	return true, nil
}
