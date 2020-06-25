package queries

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/stellar/kelp/api"
)

// sqlQueryDailyValuesTemplate queries the trades table to get the values for a given day
const sqlQueryDailyValuesTemplate = "SELECT SUM(base_volume) as total_base_volume, SUM(counter_cost) as total_counter_volume FROM trades WHERE market_id IN (%s) AND DATE(date_utc) = $1 and action = $2 group by DATE(date_utc)"

// DailyVolumeByDate is a query that fetches the daily volume of sales
type DailyVolumeByDate struct {
	db       *sql.DB
	sqlQuery string
	action   string
}

var _ api.Query = &DailyVolumeByDate{}

// DailyVolume represents any volume value which can be either bought or sold depending on the query
type DailyVolume struct {
	BaseVol  float64
	QuoteVol float64
}

// MakeDailyVolumeByDateForMarketIdsAction makes the DailyVolumeByDate query for a set of marketIds and an action
func MakeDailyVolumeByDateForMarketIdsAction(
	db *sql.DB,
	marketIDs []string,
	action string,
) (*DailyVolumeByDate, error) {
	if db == nil {
		return nil, fmt.Errorf("the provided db should be non-nil")
	}

	sqlQuery := makeSQLQueryDailyVolume(marketIDs)
	return &DailyVolumeByDate{
		db:       db,
		sqlQuery: sqlQuery,
		action:   action,
	}, nil
}

// Name impl.
func (q *DailyVolumeByDate) Name() string {
	return "DailyVolumeByDate"
}

// QueryRow impl.
func (q *DailyVolumeByDate) QueryRow(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 arg (dateUTC string), but got args %v", args)
	} else if _, ok := args[0].(string); !ok {
		return nil, fmt.Errorf("input arg needs to be of type 'string', but was of type '%T'", args[0])
	}

	row := q.db.QueryRow(q.sqlQuery, args[0], q.action)

	var baseVol sql.NullFloat64
	var quoteVol sql.NullFloat64
	e := row.Scan(&baseVol, &quoteVol)
	if e != nil {
		if strings.Contains(e.Error(), "no rows in result set") {
			return &DailyVolume{
				BaseVol:  0,
				QuoteVol: 0,
			}, nil
		}
		return nil, fmt.Errorf("could not read data from SqlQueryDailyValues query: %s", e)
	}

	if !baseVol.Valid {
		return nil, fmt.Errorf("baseVol was invalid")
	}
	if !quoteVol.Valid {
		return nil, fmt.Errorf("quoteVol was invalid")
	}

	return &DailyVolume{
		BaseVol:  baseVol.Float64,
		QuoteVol: quoteVol.Float64,
	}, nil
}

func makeSQLQueryDailyVolume(marketIDs []string) string {
	inClauseParts := []string{}
	for _, mid := range marketIDs {
		inValue := fmt.Sprintf("'%s'", mid)
		inClauseParts = append(inClauseParts, inValue)
	}
	inClause := strings.Join(inClauseParts, ", ")

	return fmt.Sprintf(sqlQueryDailyValuesTemplate, inClause)
}
