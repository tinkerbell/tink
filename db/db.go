package db

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
)

var logger log.Logger

// ConnectDB returns a connection to postgres database
func ConnectDB(lg log.Logger) *sql.DB {
	logger = lg
	db, err := sql.Open("postgres", "")
	if err != nil {
		logger.Error(err)
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

func get(ctx context.Context, db *sql.DB, query string, args ...interface{}) (string, error) {
	row := db.QueryRowContext(ctx, query, args...)

	buf := []byte{}
	err := row.Scan(&buf)
	if err == nil {
		return string(buf), nil
	}

	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT")
		logger.Error(err)
	} else {
		err = nil
	}

	return "", err
}
