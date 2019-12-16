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

var upgradeScripts = []*UpgradeScript{
	makeUpgradeScript(1, sqlDbVersionTableCreate),
	makeUpgradeScript(2,
		sqlMarketsTableCreate,
		sqlTradesTableCreate,
		sqlTradesIndexCreate,
	),
}

// UpgradeScript encapsulates a script to be run to upgrade the database from one version to the next
type UpgradeScript struct {
	version  uint32
	commands []string
}

// makeUpgradeScript encapsulates a script to be run to upgrade the database from one version to the next
func makeUpgradeScript(version uint32, command string, moreCommands ...string) *UpgradeScript {
	allCommands := []string{command}
	allCommands = append(allCommands, moreCommands...)

	return &UpgradeScript{
		version:  version,
		commands: allCommands,
	}
}

// ConnectInitializedDatabase creates a database with the required metadata tables
func ConnectInitializedDatabase(postgresDbConfig *postgresdb.Config) (*sql.DB, error) {
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
	e = runUpgradeScripts(db, upgradeScripts)
	if e != nil {
		return nil, fmt.Errorf("could not run upgrade scripts: %s", e)
	}
	log.Printf("... finished creating db schema and running upgrade scripts\n")

	return db, nil
}

func runUpgradeScripts(db *sql.DB, scripts []*UpgradeScript) error {
	currentDbVersion, e := QueryDbVersion(db)
	if e != nil {
		if !strings.Contains(e.Error(), "relation \"db_version\" does not exist") {
			return fmt.Errorf("could not fetch current db version: %s", e)
		}
		currentDbVersion = 0
	}

	for _, script := range scripts {
		if script.version <= currentDbVersion {
			log.Printf("   skipping upgrade script for version %d because current db version (%d) is equal or ahead\n", script.version, currentDbVersion)
			continue
		}

		// start transaction
		e = postgresdb.ExecuteStatement(db, "BEGIN")
		if e != nil {
			return fmt.Errorf("could not start transaction before upgrading db to version %d: %s", script.version, e)
		}

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

		// add entry to db_version table
		sqlInsertDbVersion := fmt.Sprintf(sqlDbVersionTableInsertTemplate,
			script.version,
			startTime.Format(postgresdb.DateFormatString),
			len(script.commands),
			elapsedMillis,
		)
		_, e = db.Exec(sqlInsertDbVersion)
		if e != nil {
			// duplicate insert should return an error
			return fmt.Errorf("could not execute sql insert values statement in db_version table for db version %d (%s): %s", script.version, sqlInsertDbVersion, e)
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
