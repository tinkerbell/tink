package db

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
)

// ConnectDB returns a connection to postgres database
func ConnectDB(logger log.Logger) *sql.DB {
	db, err := sql.Open("postgres", "")
	if err != nil {
		logger.Error(err)
		panic(err)
	}
	if err := truncate(db); err != nil {
		if pqErr := Error(err); pqErr != nil {
			logger.With("detail", pqErr.Detail, "where", pqErr.Where).Error(err)
		}
		panic(err)
	}
	return db
}

// Error returns the underlying cause for error
func Error(err error) *pq.Error {
	if pqErr, ok := errors.Cause(err).(*pq.Error); ok {
		return pqErr
	}
	return nil
}

func truncate(db *sql.DB) error {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec("TRUNCATE hardware")
	if err != nil {
		return errors.Wrap(err, "TRUNCATE")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "TRUNCATE")
	}
	return err
}
