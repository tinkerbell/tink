package internal

import (
	"database/sql"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/db"
)

// SetupPostgres initializes a connection to a postgres database.
func SetupPostgres(connInfo string, onlyMigrate bool, logger log.Logger) (db.Database, error) {
	dbCon, err := sql.Open("postgres", connInfo)
	if err != nil {
		return nil, err
	}
	tinkDB := db.Connect(dbCon, logger)

	if onlyMigrate {
		logger.Info("Applying migrations. This process will end when migrations will take place.")
		numAppliedMigrations, err := tinkDB.Migrate()
		if err != nil {
			return nil, err
		}
		logger.With("num_applied_migrations", numAppliedMigrations).Info("Migrations applied successfully")
		return nil, nil
	}

	numAvailableMigrations, err := tinkDB.CheckRequiredMigrations()
	if err != nil {
		return nil, err
	}
	if numAvailableMigrations != 0 {
		logger.Info("Your database schema is not up to date. Please apply migrations running tink-server with env var ONLY_MIGRATION set.")
	}
	return *tinkDB, nil
}
