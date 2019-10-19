package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// CreateTemplate create a workflow template in database
func CreateTemplate(ctx context.Context, db *sql.DB, name string, data []byte) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	INSERT INTO
		template (inserted_at, id, name, data)
	VALUES
		($1, md5($2::varchar)::uuid, $2, $3)
	ON CONFLICT (id)
	DO
	UPDATE SET
		(inserted_at, deleted_at, name, data) = ($1, NULL, $2, $3);
	`, time.Now(), name, data)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}
