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
	e := RunUpgradeScripts(db, UpgradeScripts, codeVersionString)
	if e != nil {
		panic(e)
	}

	// assert current state of the database
	assert.Equal(t, 1, GetNumTablesInDb(db))
	assert.True(t, CheckTableExists(db, "db_version"))

	// check schema of db_version table
	columns := GetTableSchema(db, "db_version")
	assert.Equal(t, 5, len(columns), fmt.Sprintf("%v", columns))
	AssertTableColumnsEqual(t, &TableColumn{
		ColumnName:             "version",
		OrdinalPosition:        1,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "integer",
		CharacterMaximumLength: nil,
	}, &columns[0])
	AssertTableColumnsEqual(t, &TableColumn{
		ColumnName:             "date_completed_utc",
		OrdinalPosition:        2,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "timestamp without time zone",
		CharacterMaximumLength: nil,
	}, &columns[1])
	AssertTableColumnsEqual(t, &TableColumn{
		ColumnName:             "num_scripts",
		OrdinalPosition:        3,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "integer",
		CharacterMaximumLength: nil,
	}, &columns[2])
	AssertTableColumnsEqual(t, &TableColumn{
		ColumnName:             "time_elapsed_millis",
		OrdinalPosition:        4,
		ColumnDefault:          nil,
		IsNullable:             "NO",
		DataType:               "bigint",
		CharacterMaximumLength: nil,
	}, &columns[3])
	AssertTableColumnsEqual(t, &TableColumn{
		ColumnName:             "code_version_string",
		OrdinalPosition:        5,
		ColumnDefault:          nil,
		IsNullable:             "YES",
		DataType:               "text",
		CharacterMaximumLength: nil,
	}, &columns[4])

	// check indexes of db_version table
	indexes := GetTableIndexes(db, "db_version")
	assert.Equal(t, 1, len(indexes))
	AssertIndex(t, "db_version", "db_version_pkey", "CREATE UNIQUE INDEX db_version_pkey ON public.db_version USING btree (version)", indexes)

	// check entries of db_version table
	allRows := QueryAllRows(db, "db_version")
	assert.Equal(t, 2, len(allRows))
	// first code_version_string is nil becuase the field was not supported at the time when the upgrade script was run, and only in version 2 of
	// the database do we add the field. See UpgradeScripts and RunUpgradeScripts() for more details
	ValidateDBVersionRow(t, allRows[0], 1, time.Now(), 1, 10, nil)
	ValidateDBVersionRow(t, allRows[1], 2, time.Now(), 1, 10, &codeVersionString)
}
