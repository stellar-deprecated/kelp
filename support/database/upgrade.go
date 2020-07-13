package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stellar/kelp/support/postgresdb"
	"github.com/stellar/kelp/support/utils"
)

// sqlDbVersionTableInsertTemplate inserts into the db_version table
const sqlDbVersionTableInsertTemplate1 = "INSERT INTO db_version (version, date_completed_utc, num_scripts, time_elapsed_millis) VALUES (%d, '%s', %d, %d)"
const sqlDbVersionTableInsertTemplate2 = "INSERT INTO db_version (version, date_completed_utc, num_scripts, time_elapsed_millis, code_version_string) VALUES (%d, '%s', %d, %d, '%s')"

// UpgradeScript encapsulates a script to be run to upgrade the database from one version to the next
type UpgradeScript struct {
	version  uint32
	commands []string
}

// MakeUpgradeScript encapsulates a script to be run to upgrade the database from one version to the next
func MakeUpgradeScript(version uint32, command string, moreCommands ...string) *UpgradeScript {
	allCommands := []string{command}
	allCommands = append(allCommands, moreCommands...)

	return &UpgradeScript{
		version:  version,
		commands: allCommands,
	}
}

var UpgradeScripts = []*UpgradeScript{
	MakeUpgradeScript(1, SqlDbVersionTableCreate),
	MakeUpgradeScript(2, SqlDbVersionTableAlter1),
}

// ConnectInitializedDatabase creates a database with the required metadata tables
func ConnectInitializedDatabase(postgresDbConfig *postgresdb.Config, upgradeScripts []*UpgradeScript, codeVersionString string) (*sql.DB, error) {
	dbCreated, e := postgresdb.CreateDatabaseIfNotExists(postgresDbConfig)
	if e != nil {
		if strings.Contains(e.Error(), "connect: connection refused") {
			utils.PrintErrorHintf("ensure your postgres database is available on %s:%d, or remove the 'POSTGRES_DB' config from your trader config file\n", postgresDbConfig.GetHost(), postgresDbConfig.GetPort())
		}
		return nil, fmt.Errorf("error when creating database from config (%+v), created=%v: %s", *postgresDbConfig, dbCreated, e)
	}
	if dbCreated {
		log.Printf("created database '%s'", postgresDbConfig.GetDbName())
	} else {
		log.Printf("did not create db '%s' because it already exists", postgresDbConfig.GetDbName())
	}

	db, e := sql.Open("postgres", postgresDbConfig.MakeConnectString())
	if e != nil {
		return nil, fmt.Errorf("could not open database: %s", e)
	}
	// don't defer db.Close() here becuase we want it open for the life of the application for now

	log.Printf("creating db schema and running upgrade scripts ...\n")
	e = RunUpgradeScripts(db, upgradeScripts, codeVersionString)
	if e != nil {
		return nil, fmt.Errorf("could not run upgrade scripts: %s", e)
	}
	log.Printf("... finished creating db schema and running upgrade scripts\n")

	return db, nil
}

// RunUpgradeScripts is a utility function that can be run from outside this package so we need to export it
func RunUpgradeScripts(db *sql.DB, scripts []*UpgradeScript, codeVersionString string) error {
	// save feature flags for the db_version table here
	hasCodeVersionString := false

	for _, script := range scripts {
		// fetch the db version inside the for loop because it constantly gets updated
		currentDbVersion, e := QueryDbVersion(db)
		if e != nil {
			if !strings.Contains(e.Error(), "relation \"db_version\" does not exist") {
				return fmt.Errorf("could not fetch current db version: %s", e)
			}
			currentDbVersion = 0
			log.Printf("this is the first run since we don't have the db_version table, so set currentDbVersion to 0\n")
		}

		if script.version <= currentDbVersion {
			log.Printf("   skipping upgrade script for version %d because current db version (%d) is equal or ahead\n", script.version, currentDbVersion)
			continue
		}

		// start transaction
		e = postgresdb.ExecuteStatement(db, "BEGIN")
		if e != nil {
			return fmt.Errorf("could not start transaction before upgrading db to version %d: %s", script.version, e)
		}
		// issue a ROLLBACK command to handle the case of the transaction failing. it's a noop if the transaction commits successfully
		defer func() {
			postgresdb.ExecuteStatement(db, "ROLLBACK")
		}()

		startTime := time.Now()
		startTimeMillis := startTime.UnixNano() / int64(time.Millisecond)
		for ci, command := range script.commands {
			e = postgresdb.ExecuteStatement(db, command)
			if e != nil {
				return fmt.Errorf("could not execute sql statement at index %d for db version %d (%s): %s", ci, script.version, command, e)
			}
			log.Printf("   executed sql statement at index %d for db version %d", ci, script.version)
		}
		endTimeMillis := time.Now().UnixNano() / int64(time.Millisecond)
		elapsedMillis := endTimeMillis - startTimeMillis

		// update feature flags here where required after running a script so we don't need to hard-code version numbers which can be different for different consumers of this API
		for _, command := range script.commands {
			if command == SqlDbVersionTableAlter1 {
				// if we have run this alter table command it means the database version has the code_version_string feature
				hasCodeVersionString = true
			}
		}

		// add entry to db_version table
		sqlInsertDbVersion := fmt.Sprintf(sqlDbVersionTableInsertTemplate1,
			script.version,
			startTime.Format(postgresdb.TimestampFormatString),
			len(script.commands),
			elapsedMillis,
		)
		if hasCodeVersionString {
			sqlInsertDbVersion = fmt.Sprintf(sqlDbVersionTableInsertTemplate2,
				script.version,
				startTime.Format(postgresdb.TimestampFormatString),
				len(script.commands),
				elapsedMillis,
				codeVersionString,
			)
		}
		_, e = db.Exec(sqlInsertDbVersion)
		if e != nil {
			// duplicate insert should return an error
			return fmt.Errorf("could not add an entry to the db_version table for upgrade script (db_version=%d) for current db version %d (%s): %s", script.version, currentDbVersion, sqlInsertDbVersion, e)
		}

		// commit transaction
		e = postgresdb.ExecuteStatement(db, "COMMIT")
		if e != nil {
			return fmt.Errorf("could not commit transaction before upgrading db to version %d: %s", script.version, e)
		}
		log.Printf("   successfully ran %d upgrade commands and upgraded to version %d of the database in %d milliseconds\n", len(script.commands), script.version, elapsedMillis)
	}
	return nil
}
