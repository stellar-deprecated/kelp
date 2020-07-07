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

func TestCurrentClassTestInfra(t *testing.T) {
	// run the preTest
	db, dbname := preTest(t)
	// assert state of preTest
	assert.NotNil(t, db)
	assert.Equal(t, strings.ToLower(dbname), dbname)
	assert.True(t, checkDatabaseExistsWithNewConnection(dbname))
	// assert that there are no tables in the database we just created
	tablesQueryResult, e := db.Query("select * from pg_stat_user_tables")
	assert.NoError(t, e)
	assert.False(t, tablesQueryResult.Next()) // this should have 0 rows
	tablesQueryResult.Close()                 // remembering to close the query, otherwise the connection will remain open in the postTestWithDbClose function

	// run the postTest
	postTestWithDbClose(db, dbname)
	// assert state after the postTest
	assert.False(t, checkDatabaseExistsWithNewConnection(dbname))
}
