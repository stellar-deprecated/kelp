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

// PreTest logic for tests related to db upgrades
func PreTest(t *testing.T) (*sql.DB, string) {
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

// PostTestWithDbClose logic for tests related to db upgrades
func PostTestWithDbClose(db *sql.DB, dbname string) {
	// defer statements are executed in LIFO order

	// second delete the database (internally creates a new db connection and then deletes it)
	defer dropDatabaseWithNewConnection(dbname)

	// first close the existing db connection
	defer db.Close()
}

// CheckDatabaseExistsWithNewConnection is well-named
func CheckDatabaseExistsWithNewConnection(dbname string) bool {
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

// GetNumTablesInDb is well-named
func GetNumTablesInDb(db *sql.DB) int {
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

// CheckTableExists is well-named
func CheckTableExists(db *sql.DB, tableName string) bool {
	tablesQueryResult, e := db.Query(fmt.Sprintf("select tablename from pg_catalog.pg_tables where tablename = '%s'", tableName))
	if e != nil {
		panic(e)
	}
	defer tablesQueryResult.Close() // remembering to defer closing the query

	return tablesQueryResult.Next()
}

// TableColumn is a column in a table specefied generically
type TableColumn struct {
	ColumnName             string      // `db:"column_name"`
	OrdinalPosition        int         // `db:"ordinal_position"`
	ColumnDefault          interface{} // `db:"column_default"`
	IsNullable             string      // `db:"is_nullable"`        // uses "YES" / "NO" instead of a boolean
	DataType               string      // `db:"data_type"`
	CharacterMaximumLength interface{} // `db:"character_maximum_length"`
}

// AssertTableColumnsEqual is well-named
func AssertTableColumnsEqual(t *testing.T, want *TableColumn, actual *TableColumn) {
	assert.Equal(t, want.ColumnName, actual.ColumnName)
	assert.Equal(t, want.OrdinalPosition, actual.OrdinalPosition)
	assert.Equal(t, want.ColumnDefault, actual.ColumnDefault)
	assert.Equal(t, want.IsNullable, actual.IsNullable)
	assert.Equal(t, want.DataType, actual.DataType)
	assert.Equal(t, want.CharacterMaximumLength, actual.CharacterMaximumLength)
}

// GetTableSchema is well-named
func GetTableSchema(db *sql.DB, tableName string) []TableColumn {
	schemaQueryResult, e := db.Query(fmt.Sprintf("SELECT column_name, ordinal_position, column_default, is_nullable, data_type, character_maximum_length FROM information_schema.columns WHERE table_schema = 'public' AND table_name = '%s'", tableName))
	if e != nil {
		panic(e)
	}
	defer schemaQueryResult.Close() // remembering to defer closing the query

	items := []TableColumn{}
	for schemaQueryResult.Next() { // remembering to call Next() before Scan()
		var item TableColumn
		e = schemaQueryResult.Scan(&item.ColumnName, &item.OrdinalPosition, &item.ColumnDefault, &item.IsNullable, &item.DataType, &item.CharacterMaximumLength)
		if e != nil {
			panic(e)
		}

		items = append(items, item)
	}

	return items
}

// IndexSearchResult captures the result from GetTableIndexes() and is used as input to AssertIndex()
type IndexSearchResult map[string]string

// GetTableIndexes is well-named
func GetTableIndexes(db *sql.DB, tableName string) IndexSearchResult {
	indexQueryResult, e := db.Query(fmt.Sprintf("SELECT indexname, indexdef from pg_indexes where schemaname = 'public' AND tablename = '%s'", tableName))
	if e != nil {
		panic(e)
	}
	defer indexQueryResult.Close() // remembering to defer closing the query

	m := map[string]string{}
	for indexQueryResult.Next() { // remembering to call Next() before Scan()
		var name, def string
		e = indexQueryResult.Scan(&name, &def)
		if e != nil {
			panic(e)
		}

		m[name] = def
	}

	return m
}

// AssertIndex validates that the index exists
func AssertIndex(t *testing.T, tableName string, wantIndexName string, wantDefinition string, indexes IndexSearchResult) {
	m := map[string]string(indexes)
	if v, ok := m[wantIndexName]; assert.True(t, ok, fmt.Sprintf("index '%s' should exist in the table '%s'", wantIndexName, tableName)) {
		assert.Equal(t, wantDefinition, v)
	}
}

// QueryAllRows queries all the rows of a given table in a database
func QueryAllRows(db *sql.DB, tableName string) [][]interface{} {
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

// ValidateDBVersionRow is well-named
func ValidateDBVersionRow(
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

func dropDatabaseWithNewConnection(dbname string) {
	execWithNewManagedConnection(func(db *sql.DB) {
		_, e := db.Exec(fmt.Sprintf("DROP DATABASE %s", dbname))
		if e != nil {
			panic(e)
		}
	})
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
