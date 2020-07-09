package database

import (
	"database/sql"
	"fmt"
	"log"
)

/*
	tables
*/
const SqlDbVersionTableCreate = "CREATE TABLE IF NOT EXISTS db_version (version INTEGER NOT NULL, date_completed_utc TIMESTAMP WITHOUT TIME ZONE NOT NULL, num_scripts INTEGER NOT NULL, time_elapsed_millis BIGINT NOT NULL, PRIMARY KEY (version))"
const SqlDbVersionTableAlter1 = "ALTER TABLE db_version ADD COLUMN code_version_string TEXT"

/*
	queries
*/
// sqlQueryDbVersion queries the db_version table
const sqlQueryDbVersion = "SELECT version FROM db_version ORDER BY version desc LIMIT 1"

/*
	query helper functions
*/
// QueryDbVersion queries for the version of the database
func QueryDbVersion(db *sql.DB) (uint32, error) {
	rows, e := db.Query(sqlQueryDbVersion)
	if e != nil {
		return 0, fmt.Errorf("could not execute sql select query (%s): %s", sqlQueryDbVersion, e)
	}
	defer rows.Close()

	for rows.Next() {
		var dbVersion uint32
		e = rows.Scan(&dbVersion)
		if e != nil {
			return 0, fmt.Errorf("could not scan row to get the db version: %s", e)
		}

		log.Printf("fetched dbVersion from db: %d", dbVersion)
		return dbVersion, nil
	}

	return 0, nil
}
