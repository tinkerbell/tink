package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// InsertIntoTargetDB : Push targets data in target table
func InsertIntoTargetDB(ctx context.Context, db *sql.DB, data string, uuid string) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
        INSERT INTO
                targets (inserted_at, id, data)
        VALUES
                ($1, $2, $3)
        ON CONFLICT (id)
        DO
        UPDATE SET
                (inserted_at, deleted_at, data) = ($1, NULL, $3);
        `, time.Now(), uuid, data)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// TargetsByID : Get the targets data which belongs to the input id
func TargetsByID(ctx context.Context, db *sql.DB, id string) (string, error) {
	arg := id

	query := `
	SELECT data
	FROM targets
	WHERE
		deleted_at IS NULL
	AND
		id = $1
	`
	return get(ctx, db, query, arg)
}

// DeleteFromTargetDB : Delete the targets which belong to the input id
func DeleteFromTargetDB(ctx context.Context, db *sql.DB, id string) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	UPDATE targets
	SET
		deleted_at = NOW()
	WHERE
		id = $1;
	`, id)

	if err != nil {
		return errors.Wrap(err, "DELETE")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// ListTargets returns all saved targets which are not deleted
func ListTargets(db *sql.DB, fn func(id, n string) error) error {
	rows, err := db.Query(`
        SELECT id, data
        FROM targets
        WHERE
                deleted_at IS NULL;
        `)

	if err != nil {
		return err
	}

	defer rows.Close()
	var (
		id   string
		data string
	)

	for rows.Next() {
		err = rows.Scan(&id, &data)
		if err != nil {
			err = errors.Wrap(err, "SELECT")
			logger.Error(err)
			return err
		}
		err = fn(id, data)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if err == sql.ErrNoRows {
		err = nil
	}
	return err
}
