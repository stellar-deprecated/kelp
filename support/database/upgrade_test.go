package database

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/support/postgresdb"
)

func preTest(t *testing.T) (*sql.DB, string) {
	// use ToLower because the database name cannot be uppercase in postgres
	dbname := fmt.Sprintf("test_database_%s_%d", strings.ToLower(t.Name()), time.Now().UnixNano())
	postgresDbConfig := &postgresdb.Config{
		Host:      "localhost",
		Port:      5432,
		DbName:    dbname,
		User:      os.Getenv("POSTGRES_USER"),
		SSLEnable: false,
	}

	// create empty database
	db, e := ConnectInitializedDatabase(postgresDbConfig, []*UpgradeScript{}, "")
	if e != nil {
		panic(e)
	}

	return db, dbname
}

// execWithNewManagedConnection creates a new connection which is "managed" (i.e. closing is handled) so the passed in function can just use it to do their business
func execWithNewManagedConnection(fn func(db *sql.DB)) {
	postgresDbConfig := &postgresdb.Config{
		Host:      "localhost",
		Port:      5432,
		DbName:    "postgres",
		User:      os.Getenv("POSTGRES_USER"),
		SSLEnable: false,
	}
	// connect to the db
	db, e := sql.Open("postgres", postgresDbConfig.MakeConnectStringWithoutDB())
	if e != nil {
		panic(e)
	}
	// defer closing this new connection
	defer db.Close()

	// delegate to passed in function
	fn(db)
}

func dropDatabaseWithNewConnection(dbname string) {
	execWithNewManagedConnection(func(db *sql.DB) {
		_, e := db.Exec(fmt.Sprintf("DROP DATABASE %s", dbname))
		if e != nil {
			panic(e)
		}
	})
}

func postTestWithDbClose(db *sql.DB, dbname string) {
	// defer statements are executed in LIFO order

	// second delete the database (internally creates a new db connection and then deletes it)
	defer dropDatabaseWithNewConnection(dbname)

	// first close the existing db connection
	defer db.Close()
}

func checkDatabaseExistsWithNewConnection(dbname string) bool {
	hasDatabase := false
	execWithNewManagedConnection(func(db *sql.DB) {
		rows, e := db.Query(fmt.Sprintf("SELECT datname FROM pg_database WHERE datname = '%s'", dbname))
		if e != nil {
			panic(e)
		}

		hasDatabase = rows.Next()
	})
	return hasDatabase
}

func getNumTablesInDb(db *sql.DB) int {
	// run the query -- note that we need to be connected to the database of interest
	tablesQueryResult, e := db.Query("select COUNT(*) from pg_stat_user_tables")
	if e != nil {
		panic(e)
	}
	defer tablesQueryResult.Close() // remembering to defer closing the query

	tablesQueryResult.Next() // remembering to call Next() before Scan()
	var count int
	e = tablesQueryResult.Scan(&count)
	if e != nil {
		panic(e)
	}

	return count
}

func checkTableExists(db *sql.DB, tableName string) bool {
	tablesQueryResult, e := db.Query(fmt.Sprintf("select tablename from pg_catalog.pg_tables where tablename = '%s'", tableName))
	if e != nil {
		panic(e)
	}
	defer tablesQueryResult.Close() // remembering to defer closing the query

	return tablesQueryResult.Next()
}

type tableColumn struct {
	columnName             string      // `db:"column_name"`
	ordinalPosition        int         // `db:"ordinal_position"`
	columnDefault          interface{} // `db:"column_default"`
	isNullable             string      // `db:"is_nullable"`        // uses "YES" / "NO" instead of a boolean
	dataType               string      // `db:"data_type"`
	characterMaximumLength interface{} // `db:"character_maximum_length"`
}

func assertTableColumnsEqual(t *testing.T, want *tableColumn, actual *tableColumn) {
	assert.Equal(t, want.columnName, actual.columnName)
	assert.Equal(t, want.ordinalPosition, actual.ordinalPosition)
	assert.Equal(t, want.columnDefault, actual.columnDefault)
	assert.Equal(t, want.isNullable, actual.isNullable)
	assert.Equal(t, want.dataType, actual.dataType)
	assert.Equal(t, want.characterMaximumLength, actual.characterMaximumLength)
}

func getTableSchema(db *sql.DB, tableName string) []tableColumn {
	schemaQueryResult, e := db.Query(fmt.Sprintf("SELECT column_name, ordinal_position, column_default, is_nullable, data_type, character_maximum_length FROM information_schema.columns WHERE table_schema = 'public' AND table_name = '%s'", tableName))
	if e != nil {
		panic(e)
	}
	defer schemaQueryResult.Close() // remembering to defer closing the query

	items := []tableColumn{}
	for schemaQueryResult.Next() { // remembering to call Next() before Scan()
		var item tableColumn
		e = schemaQueryResult.Scan(&item.columnName, &item.ordinalPosition, &item.columnDefault, &item.isNullable, &item.dataType, &item.characterMaximumLength)
		if e != nil {
			panic(e)
		}

		items = append(items, item)
	}

	return items
}

func queryAllRows(db *sql.DB, tableName string) [][]interface{} {
	queryResult, e := db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if e != nil {
		panic(e)
	}
	defer queryResult.Close() // remembering to defer closing the query

	allRows := [][]interface{}{}
	for queryResult.Next() { // remembering to call Next() before Scan()
		// we want to generically query the table
		colTypes, e := queryResult.ColumnTypes()
		if e != nil {
			panic(e)
		}
		columnValues := []interface{}{}
		for i := 0; i < len(colTypes); i++ {
			columnValues = append(columnValues, new(interface{}))
		}

		e = queryResult.Scan(columnValues...)
		if e != nil {
			panic(e)
		}

		allRows = append(allRows, columnValues)
	}

	return allRows
}

func TestCurrentClassTestInfra(t *testing.T) {
	// run the preTest
	db, dbname := preTest(t)
	// assert state of preTest
	assert.NotNil(t, db)
	assert.Equal(t, strings.ToLower(dbname), dbname)
	assert.True(t, checkDatabaseExistsWithNewConnection(dbname))
	assert.Equal(t, 0, getNumTablesInDb(db))

	// run the postTest
	postTestWithDbClose(db, dbname)
	// assert state after the postTest
	assert.False(t, checkDatabaseExistsWithNewConnection(dbname))
}

func TestUpgradeScripts(t *testing.T) {
	// run the preTest and defer running the postTest
	db, dbname := preTest(t)
	defer postTestWithDbClose(db, dbname)

	// run the upgrade scripts
	codeVersionString := "someCodeVersion"
	runUpgradeScripts(db, UpgradeScripts, codeVersionString)

	// assert current state of the database
	assert.Equal(t, 1, getNumTablesInDb(db))
	assert.True(t, checkTableExists(db, "db_version"))

	// check schema of db_version table
	columns := getTableSchema(db, "db_version")
	assert.Equal(t, 5, len(columns), fmt.Sprintf("%v", columns))
	assertTableColumnsEqual(t, &tableColumn{
		columnName:             "version",
		ordinalPosition:        1,
		columnDefault:          nil,
		isNullable:             "NO",
		dataType:               "integer",
		characterMaximumLength: nil,
	}, &columns[0])
	assertTableColumnsEqual(t, &tableColumn{
		columnName:             "date_completed_utc",
		ordinalPosition:        2,
		columnDefault:          nil,
		isNullable:             "NO",
		dataType:               "timestamp without time zone",
		characterMaximumLength: nil,
	}, &columns[1])
	assertTableColumnsEqual(t, &tableColumn{
		columnName:             "num_scripts",
		ordinalPosition:        3,
		columnDefault:          nil,
		isNullable:             "NO",
		dataType:               "integer",
		characterMaximumLength: nil,
	}, &columns[2])
	assertTableColumnsEqual(t, &tableColumn{
		columnName:             "time_elapsed_millis",
		ordinalPosition:        4,
		columnDefault:          nil,
		isNullable:             "NO",
		dataType:               "bigint",
		characterMaximumLength: nil,
	}, &columns[3])
	assertTableColumnsEqual(t, &tableColumn{
		columnName:             "code_version_string",
		ordinalPosition:        5,
		columnDefault:          nil,
		isNullable:             "YES",
		dataType:               "text",
		characterMaximumLength: nil,
	}, &columns[4])

	// check entries of db_version table
	allRows := queryAllRows(db, "db_version")
	assert.Equal(t, 2, len(allRows))
	// first code_version_string is nil becuase the field was not supported at the time when the upgrade script was run, and only in version 2 of
	// the database do we add the field. See UpgradeScripts and runUpgradeScripts() for more details
	validateDBVersionRow(t, allRows[0], 1, time.Now(), 1, 10, nil)
	validateDBVersionRow(t, allRows[1], 2, time.Now(), 1, 10, &codeVersionString)
}

func validateDBVersionRow(
	t *testing.T,
	actualRow []interface{},
	wantVersion int,
	wantDateCompletedUTC time.Time,
	wantNumScripts int,
	wantTimeElapsedMillis int,
	wantCodeVersionString *string,
) {
	// first check length
	if assert.Equal(t, 5, len(actualRow)) {
		assert.Equal(t, fmt.Sprintf("%d", wantVersion), fmt.Sprintf("%v", reflect.ValueOf(actualRow[0]).Elem()))
		assert.Equal(t, wantDateCompletedUTC.Format("20060102"), reflect.ValueOf(actualRow[1]).Elem().Interface().(time.Time).Format("20060102"))
		assert.Equal(t, fmt.Sprintf("%v", wantNumScripts), fmt.Sprintf("%v", reflect.ValueOf(actualRow[2]).Elem()))
		elapsed, e := strconv.Atoi(fmt.Sprintf("%v", reflect.ValueOf(actualRow[3]).Elem()))
		if assert.Nil(t, e) {
			assert.LessOrEqual(t, elapsed, wantTimeElapsedMillis)
		}
		if wantCodeVersionString == nil {
			assert.Equal(t, "<nil>", fmt.Sprintf("%v", reflect.ValueOf(actualRow[4]).Elem()))
		} else {
			assert.Equal(t, *wantCodeVersionString, fmt.Sprintf("%v", reflect.ValueOf(actualRow[4]).Elem()))
		}
	}
}
