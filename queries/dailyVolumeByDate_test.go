package queries

import (
	"database/sql"
	"fmt"
	"os"
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
	// setup db
	yesterday, _ := time.Parse(time.RFC3339, "2020-01-20T15:00:00Z")
	today, _ := time.Parse(time.RFC3339, "2020-01-21T15:00:00Z")
	tomorrow, _ := time.Parse(time.RFC3339, "2020-01-22T15:00:00Z")
	setupStatements := []string{
		kelpdb.SqlTradesTableCreate,
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

	// make query being tested
	dailyVolumeByDateQuery, e := MakeDailyVolumeByDateForMarketIdsAction(
		db,
		[]string{"market1"},
		"sell",
	)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, "DailyVolumeByDate", dailyVolumeByDateQuery.Name())

	runQueryAndVerifyValues(t, dailyVolumeByDateQuery, yesterday, 100.0, 10.0)
	runQueryAndVerifyValues(t, dailyVolumeByDateQuery, today, 107.0, 11.83)
	runQueryAndVerifyValues(t, dailyVolumeByDateQuery, tomorrow, 102.0, 12.24)
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
