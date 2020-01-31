package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// DeleteFromDB : delete data from hardware table
func DeleteFromDB(ctx context.Context, db *sql.DB, id string) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	UPDATE hardware
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

// InsertIntoDB : insert data into hardware table
func InsertIntoDB(ctx context.Context, db *sql.DB, data string) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	INSERT INTO
		hardware (inserted_at, id, data)
	VALUES
		($1, ($2::jsonb ->> 'id')::uuid, $2)
	ON CONFLICT (id)
	DO
	UPDATE SET
		(inserted_at, deleted_at, data) = ($1, NULL, $2);
	`, time.Now(), data)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// GetByMAC : get data by machine mac
func GetByMAC(ctx context.Context, db *sql.DB, mac string) (string, error) {
	arg := `
	{
	  "network_ports": [
	    {
	      "data": {
		"mac": "` + mac + `"
	      }
	    }
	  ]
	}
	`
	query := `
	SELECT data
	FROM hardware
	WHERE
		deleted_at IS NULL
	AND
		data @> $1
	`

	return get(ctx, db, query, arg)
}

// GetByIP : get data by machine ip
func GetByIP(ctx context.Context, db *sql.DB, ip string) (string, error) {
	instance := `
	{
	  "instance": {
	    "ip_addresses": [
	      {
		"address": "` + ip + `"
	      }
	    ]
	  }
	}
	`
	hardwareOrManagement := `
	{
		"ip_addresses": [
			{
				"address": "` + ip + `"
			}
		]
	}
	`

	query := `
	SELECT data
	FROM hardware
	WHERE
		deleted_at IS NULL
	AND (
		data @> $1
		OR
		data @> $2
	)
	`

	return get(ctx, db, query, instance, hardwareOrManagement)
}

// GetByID : get data by machine id
func GetByID(ctx context.Context, db *sql.DB, id string) (string, error) {
	arg := id

	query := `
	SELECT data
	FROM hardware
	WHERE
		deleted_at IS NULL
	AND
		id = $1
	`
	return get(ctx, db, query, arg)
}

// GetAll : get data for all machine
func GetAll(db *sql.DB, fn func(string) error) error {
	rows, err := db.Query(`
	SELECT data
	FROM hardware
	WHERE
		deleted_at IS NULL
	`)

	if err != nil {
		return err
	}

	defer rows.Close()
	buf := []byte{}
	for rows.Next() {
		err = rows.Scan(&buf)
		if err != nil {
			err = errors.Wrap(err, "SELECT")
			logger.Error(err)
			return err
		}

		err = fn(string(buf))
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
