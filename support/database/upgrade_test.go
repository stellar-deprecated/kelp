package database

import (
	"database/sql"
	"fmt"
	"os"
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
	db, e := ConnectInitializedDatabase(postgresDbConfig, []*UpgradeScript{})
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
	runUpgradeScripts(db, UpgradeScripts)

	// assert current state of the database
	assert.Equal(t, 1, getNumTablesInDb(db))
	assert.True(t, checkTableExists(db, "db_version"))
	// TODO check schema of db_version table
	// TODO check entries of db_version table
}
