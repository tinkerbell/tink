package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteFromDB : delete data from hardware table.
func (d TinkDB) DeleteFromDB(ctx context.Context, id string) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	res, err := tx.Exec(`
	UPDATE hardware
	SET
		deleted_at = NOW()
	WHERE
		id = $1;
	`, id)
	if err != nil {
		return errors.Wrap(err, "DELETE")
	}

	if count, _ := res.RowsAffected(); count == int64(0) {
		return status.Error(codes.NotFound, fmt.Sprintf("not found, id:%s", id))
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// InsertIntoDB : insert data into hardware table.
func (d TinkDB) InsertIntoDB(ctx context.Context, data string) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
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

// GetByMAC : get data by machine mac.
func (d TinkDB) GetByMAC(ctx context.Context, mac string) (string, error) {
	arg := `
	{
		"network": {
			"interfaces": [
				{
					"dhcp": {
						"mac": "` + mac + `"
					}
				}
			]
		}
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

	return get(ctx, d.instance, query, arg)
}

// GetByIP : get data by machine ip.
func (d TinkDB) GetByIP(ctx context.Context, ip string) (string, error) {
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
		"network": {
			"interfaces": [
				{
					"dhcp": {
						"ip": {
							"address": "` + ip + `"
						}
					}
				}
			]
		}
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

	return get(ctx, d.instance, query, instance, hardwareOrManagement)
}

// GetByID : get data by machine id.
func (d TinkDB) GetByID(ctx context.Context, id string) (string, error) {
	arg := id

	query := `
	SELECT data
	FROM hardware
	WHERE
		deleted_at IS NULL
	AND
		id = $1
	`
	return get(ctx, d.instance, query, arg)
}

// GetAll : get data for all machine.
func (d TinkDB) GetAll(fn func([]byte) error) error {
	rows, err := d.instance.Query(`
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
			d.logger.Error(err)
			return err
		}

		err = fn(buf)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return err
}
