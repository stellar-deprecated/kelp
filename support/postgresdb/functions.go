package postgresdb

import (
	"database/sql"
	"fmt"
	"strings"
)

// TimestampFormatString is the format to be used when inserting timestamps in the database
const TimestampFormatString = "2006/01/02 15:04:05 MST"

// DateFormatString is the format to be used when converting a timestamp to a date
const DateFormatString = "2006/01/02"

// CreateDatabaseIfNotExists returns whether the db was created and an error if creation failed
func CreateDatabaseIfNotExists(postgresDbConfig *Config) (bool, error) {
	dbName := postgresDbConfig.GetDbName()
	db, e := sql.Open("postgres", postgresDbConfig.MakeConnectStringWithoutDB())
	if e != nil {
		return false, fmt.Errorf("could not connect to postgres instance: %s", e)
	}

	_, e = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if e != nil {
		if strings.Contains(e.Error(), fmt.Sprintf("database \"%s\" already exists", dbName)) {
			return false, nil
		}
		return false, fmt.Errorf("could not create database '%s': %s", dbName, e)
	}

	e = db.Close()
	if e != nil {
		return true, fmt.Errorf("could not close connection after creating database '%s': %s", dbName, e)
	}
	return true, nil
}

// ExecuteStatement runs a statement that does not return a result
func ExecuteStatement(db *sql.DB, sqlStatement string) error {
	statement, e := db.Prepare(sqlStatement)
	if e != nil {
		return fmt.Errorf("could not prepare sql statement (%s): %s", sqlStatement, e)
	}

	_, e = statement.Exec()
	if e != nil {
		return fmt.Errorf("could not execute sql statement (%s): %s", sqlStatement, e)
	}

	return nil
}
