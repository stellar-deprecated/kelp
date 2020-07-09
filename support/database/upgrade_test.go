package database

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCurrentClassTestInfra(t *testing.T) {
	// run the PreTest
	db, dbname := PreTest(t)
	// assert state of PreTest
	assert.NotNil(t, db)
	assert.Equal(t, strings.ToLower(dbname), dbname)
	assert.True(t, CheckDatabaseExistsWithNewConnection(dbname))
	assert.Equal(t, 0, GetNumTablesInDb(db))

	// run the postTest
	PostTestWithDbClose(db, dbname)
	// assert state after the postTest
	assert.False(t, CheckDatabaseExistsWithNewConnection(dbname))
}

func TestUpgradeScripts(t *testing.T) {
	// run the PreTest and defer running the postTest
	db, dbname := PreTest(t)
	defer PostTestWithDbClose(db, dbname)

	// run the upgrade scripts
	codeVersionString := "someCodeVersion"
	runUpgradeScripts(db, UpgradeScripts, codeVersionString)

	// assert current state of the database
	assert.Equal(t, 1, GetNumTablesInDb(db))
	assert.True(t, CheckTableExists(db, "db_version"))

	// check schema of db_version table
	columns := GetTableSchema(db, "db_version")
	assert.Equal(t, 5, len(columns), fmt.Sprintf("%v", columns))
	AssertTableColumnsEqual(t, &TableColumn{
		columnName:             "version",
		ordinalPosition:        1,
		columnDefault:          nil,
		isNullable:             "NO",
		dataType:               "integer",
		characterMaximumLength: nil,
	}, &columns[0])
	AssertTableColumnsEqual(t, &TableColumn{
		columnName:             "date_completed_utc",
		ordinalPosition:        2,
		columnDefault:          nil,
		isNullable:             "NO",
		dataType:               "timestamp without time zone",
		characterMaximumLength: nil,
	}, &columns[1])
	AssertTableColumnsEqual(t, &TableColumn{
		columnName:             "num_scripts",
		ordinalPosition:        3,
		columnDefault:          nil,
		isNullable:             "NO",
		dataType:               "integer",
		characterMaximumLength: nil,
	}, &columns[2])
	AssertTableColumnsEqual(t, &TableColumn{
		columnName:             "time_elapsed_millis",
		ordinalPosition:        4,
		columnDefault:          nil,
		isNullable:             "NO",
		dataType:               "bigint",
		characterMaximumLength: nil,
	}, &columns[3])
	AssertTableColumnsEqual(t, &TableColumn{
		columnName:             "code_version_string",
		ordinalPosition:        5,
		columnDefault:          nil,
		isNullable:             "YES",
		dataType:               "text",
		characterMaximumLength: nil,
	}, &columns[4])

	// check entries of db_version table
	allRows := QueryAllRows(db, "db_version")
	assert.Equal(t, 2, len(allRows))
	// first code_version_string is nil becuase the field was not supported at the time when the upgrade script was run, and only in version 2 of
	// the database do we add the field. See UpgradeScripts and runUpgradeScripts() for more details
	ValidateDBVersionRow(t, allRows[0], 1, time.Now(), 1, 10, nil)
	ValidateDBVersionRow(t, allRows[1], 2, time.Now(), 1, 10, &codeVersionString)
}
