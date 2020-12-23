package queries

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/kelpdb"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/postgresdb"
)

func connectTestDb() *sql.DB {
	postgresDbConfig := &postgresdb.Config{
		Host:      "localhost",
		Port:      5432,
		DbName:    "test_database",
		User:      os.Getenv("POSTGRES_USER"),
		SSLEnable: false,
	}

	_, e := postgresdb.CreateDatabaseIfNotExists(postgresDbConfig)
	if e != nil {
		panic(e)
	}

	db, e := sql.Open("postgres", postgresDbConfig.MakeConnectString())
	if e != nil {
		panic(e)
	}
	return db
}

func TestDailyVolumeByDate_QueryRow(t *testing.T) {
	testCases := []struct {
		action                    DailyVolumeAction
		queryByOptionalAccountIDs []string
		wantYesterdayBase         float64
		wantYesterdayQuote        float64
		wantTodayBase             float64
		wantTodayQuote            float64
		wantTomorrowBase          float64
		wantTomorrowQuote         float64
	}{
		{
			action:                    DailyVolumeActionSell,
			queryByOptionalAccountIDs: []string{}, // accountID1 and accountID2 are the only ones that exists
			wantYesterdayBase:         100.0,
			wantYesterdayQuote:        10.0,
			wantTodayBase:             207.0,
			wantTodayQuote:            21.83,
			wantTomorrowBase:          102.0,
			wantTomorrowQuote:         12.24,
		}, {
			action:                    DailyVolumeActionSell,
			queryByOptionalAccountIDs: nil,
			wantYesterdayBase:         100.0,
			wantYesterdayQuote:        10.0,
			wantTodayBase:             207.0,
			wantTodayQuote:            21.83,
			wantTomorrowBase:          102.0,
			wantTomorrowQuote:         12.24,
		}, {
			action:                    DailyVolumeActionSell,
			queryByOptionalAccountIDs: []string{"accountID1", "accountID2"}, // accountID1 and accountID2 are the only ones that exists
			wantYesterdayBase:         100.0,
			wantYesterdayQuote:        10.0,
			wantTodayBase:             207.0,
			wantTodayQuote:            21.83,
			wantTomorrowBase:          102.0,
			wantTomorrowQuote:         12.24,
		}, {
			action:                    DailyVolumeActionSell,
			queryByOptionalAccountIDs: []string{"accountID1"}, // accountID1 has most of the entries
			wantYesterdayBase:         100.0,
			wantYesterdayQuote:        10.0,
			wantTodayBase:             107.0,
			wantTodayQuote:            11.83,
			wantTomorrowBase:          102.0,
			wantTomorrowQuote:         12.24,
		}, {
			action:                    DailyVolumeActionSell,
			queryByOptionalAccountIDs: []string{"accountID2"}, //accountID2 has only one entry, which is for today
			wantYesterdayBase:         0.0,
			wantYesterdayQuote:        0.0,
			wantTodayBase:             100.0,
			wantTodayQuote:            10.0,
			wantTomorrowBase:          0.0,
			wantTomorrowQuote:         0.0,
		}, {
			action:                    DailyVolumeActionSell,
			queryByOptionalAccountIDs: []string{"accountID2", "accountID2"}, // duplicate accountIDs should return same as previous test case
			wantYesterdayBase:         0.0,
			wantYesterdayQuote:        0.0,
			wantTodayBase:             100.0,
			wantTodayQuote:            10.0,
			wantTomorrowBase:          0.0,
			wantTomorrowQuote:         0.0,
		}, {
			action:                    DailyVolumeActionSell,
			queryByOptionalAccountIDs: []string{"accountID3"}, //accountID3 does not exist
			wantYesterdayBase:         0.0,
			wantYesterdayQuote:        0.0,
			wantTodayBase:             0.0,
			wantTodayQuote:            0.0,
			wantTomorrowBase:          0.0,
			wantTomorrowQuote:         0.0,
		},
		{
			action:                    DailyVolumeActionBuy,
			queryByOptionalAccountIDs: []string{}, // accountID1 and accountID2 are the only ones that exists
			wantYesterdayBase:         105,
			wantYesterdayQuote:        9.9,
			wantTodayBase:             92,
			wantTodayQuote:            4.2,
			wantTomorrowBase:          303.0,
			wantTomorrowQuote:         20.64,
		},
		{
			action:                    DailyVolumeActionBuy,
			queryByOptionalAccountIDs: nil,
			wantYesterdayBase:         105,
			wantYesterdayQuote:        9.9,
			wantTodayBase:             92,
			wantTodayQuote:            4.2,
			wantTomorrowBase:          303.0,
			wantTomorrowQuote:         20.64,
		},
		{
			action:                    DailyVolumeActionBuy,
			queryByOptionalAccountIDs: []string{"accountID1", "accountID2"}, // accountID1 and accountID2 are the only ones that exists
			wantYesterdayBase:         105,
			wantYesterdayQuote:        9.9,
			wantTodayBase:             92,
			wantTodayQuote:            4.2,
			wantTomorrowBase:          303.0,
			wantTomorrowQuote:         20.64,
		},
		{
			action:                    DailyVolumeActionBuy,
			queryByOptionalAccountIDs: []string{"accountID1"}, //accountID1 has only one entry for tomorrow
			wantYesterdayBase:         0,
			wantYesterdayQuote:        0,
			wantTodayBase:             0,
			wantTodayQuote:            0,
			wantTomorrowBase:          102,
			wantTomorrowQuote:         12.24,
		},
		{
			action:                    DailyVolumeActionBuy,
			queryByOptionalAccountIDs: []string{"accountID2"}, // accountID2 has most of the entries
			wantYesterdayBase:         105,
			wantYesterdayQuote:        9.9,
			wantTodayBase:             92,
			wantTodayQuote:            4.2,
			wantTomorrowBase:          201,
			wantTomorrowQuote:         8.4,
		},
		{
			action:                    DailyVolumeActionBuy,
			queryByOptionalAccountIDs: []string{"accountID2", "accountID2"}, // duplicates should be the same as above
			wantYesterdayBase:         105,
			wantYesterdayQuote:        9.9,
			wantTodayBase:             92,
			wantTodayQuote:            4.2,
			wantTomorrowBase:          201,
			wantTomorrowQuote:         8.4,
		},
		{
			action:                    DailyVolumeActionBuy,
			queryByOptionalAccountIDs: []string{"accountID3"}, // accountID3 does not exist
			wantYesterdayBase:         0,
			wantYesterdayQuote:        0,
			wantTodayBase:             0,
			wantTodayQuote:            0,
			wantTomorrowBase:          0,
			wantTomorrowQuote:         0,
		},
	}

	// setup db
	yesterday, _ := time.Parse(time.RFC3339, "2020-01-20T15:00:00Z")
	today, _ := time.Parse(time.RFC3339, "2020-01-21T15:00:00Z")
	tomorrow, _ := time.Parse(time.RFC3339, "2020-01-22T15:00:00Z")
	setupStatements := []string{
		kelpdb.SqlTradesTableCreate,
		"ALTER TABLE trades DROP COLUMN IF EXISTS account_id",
		"ALTER TABLE trades DROP COLUMN IF EXISTS order_id",
		kelpdb.SqlTradesTableAlter1,
		kelpdb.SqlTradesTableAlter2,
		"DELETE FROM trades", // clear table
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"1",
			yesterday.Format(postgresdb.TimestampFormatString),
			model.OrderActionSell.String(),
			model.OrderTypeLimit.String(),
			0.10,  // price
			100.0, // volume
			10.0,  // cost
			0.0,   // fee
			"accountID1",
			"",
		),
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"2",
			today.Format(postgresdb.TimestampFormatString),
			model.OrderActionSell.String(),
			model.OrderTypeLimit.String(),
			0.11,  // price
			101.0, // volume
			11.11, // cost
			0.0,   // fee
			"accountID1",
			"oid1",
		),
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"3",
			today.Add(time.Second*1).Format(postgresdb.TimestampFormatString),
			model.OrderActionSell.String(),
			model.OrderTypeLimit.String(),
			0.12, // price
			6.0,  // volume
			0.72, // cost
			0.10, // fee
			"accountID1",
			"",
		),
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"4",
			tomorrow.Format(postgresdb.TimestampFormatString),
			model.OrderActionSell.String(),
			model.OrderTypeLimit.String(),
			0.12,  // price
			102.0, // volume
			12.24, // cost
			0.0,   // fee
			"accountID1",
			"",
		),
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"5",
			tomorrow.Add(time.Second*1).Format(postgresdb.TimestampFormatString),
			model.OrderActionBuy.String(),
			model.OrderTypeLimit.String(),
			0.12,  // price
			102.0, // volume
			12.24, // cost
			0.0,   // fee
			"accountID1",
			"",
		),
		// add an extra one for accountID2
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"6",
			today.Format(postgresdb.TimestampFormatString),
			model.OrderActionSell.String(),
			model.OrderTypeLimit.String(),
			0.10,  // price
			100.0, // volume
			10.0,  // cost
			0.0,   // fee
			"accountID2",
			"",
		),
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"7",
			yesterday.Format(postgresdb.TimestampFormatString),
			model.OrderActionBuy.String(),
			model.OrderTypeLimit.String(),
			0.20,  // price
			105.0, // volume
			9.9,   // cost
			0.0,   // fee
			"accountID2",
			"",
		),
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"8",
			today.Format(postgresdb.TimestampFormatString),
			model.OrderActionBuy.String(),
			model.OrderTypeLimit.String(),
			0.44, // price
			92.0, // volume
			4.2,  // cost
			0.5,  // fee
			"accountID2",
			"",
		),
		fmt.Sprintf(kelpdb.SqlTradesInsertTemplate,
			"market1",
			"9",
			tomorrow.Add(time.Second*1).Format(postgresdb.TimestampFormatString),
			model.OrderActionBuy.String(),
			model.OrderTypeLimit.String(),
			0.82,  // price
			201.0, // volume
			8.4,   // cost
			0.7,   // fee
			"accountID2",
			"",
		),
	}
	db := connectTestDb()
	defer db.Close()
	for _, s := range setupStatements {
		_, e := db.Exec(s)
		if e != nil {
			panic(e)
		}
	}

	for _, k := range testCases {
		t.Run(strings.Replace(fmt.Sprintf("%v/%s", k.queryByOptionalAccountIDs, k.action), " ", "_", -1), func(t *testing.T) {
			// make query being tested
			dailyVolumeByDateQuery, e := MakeDailyVolumeByDateForMarketIdsAction(
				db,
				[]string{"market1"},
				k.action,
				k.queryByOptionalAccountIDs,
			)
			if !assert.NoError(t, e) {
				return
			}
			assert.Equal(t, "DailyVolumeByDate", dailyVolumeByDateQuery.Name())

			runQueryAndVerifyValues(t, dailyVolumeByDateQuery, yesterday, k.wantYesterdayBase, k.wantYesterdayQuote)
			runQueryAndVerifyValues(t, dailyVolumeByDateQuery, today, k.wantTodayBase, k.wantTodayQuote)
			runQueryAndVerifyValues(t, dailyVolumeByDateQuery, tomorrow, k.wantTomorrowBase, k.wantTomorrowQuote)
		})
	}
}

func runQueryAndVerifyValues(t *testing.T, query api.Query, inputDate time.Time, wantBaseVol float64, wantQuoteVol float64) {
	result, e := query.QueryRow(inputDate.Format(postgresdb.DateFormatString))
	if e != nil {
		panic(e)
	}

	dailyVolume, ok := result.(*DailyVolume)
	if !assert.True(t, ok) {
		return
	}

	assert.Equal(t, wantBaseVol, dailyVolume.BaseVol)
	assert.Equal(t, wantQuoteVol, dailyVolume.QuoteVol)
}
