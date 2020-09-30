package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	wflow "github.com/tinkerbell/tink/workflow"
)

// CreateTemplate creates a new workflow template
func (d TinkDB) CreateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	wf, err := wflow.ParseYAML([]byte(data))
	if err != nil {
		return err
	}
	err = wflow.ValidateTemplate(wf)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	INSERT INTO
		template (created_at, updated_at, name, data, id)
	VALUES
		($1, $1, $2, $3, $4)
	ON CONFLICT (id)
	DO
	UPDATE SET
		(updated_at, deleted_at, name, data) = ($1, NULL, $2, $3);
	`, time.Now(), name, data, id)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// GetTemplate returns a workflow template
func (d TinkDB) GetTemplate(ctx context.Context, id string) (string, string, error) {
	query := `
	SELECT name, data
	FROM template
	WHERE
		id = $1
	AND
		deleted_at IS NULL
	`
	row := d.instance.QueryRowContext(ctx, query, id)
	name := []byte{}
	data := []byte{}
	err := row.Scan(&name, &data)
	if err == nil {
		return string(name), string(data), nil
	}
	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT")
		logger.Error(err)
	}
	return "", "", err
}

// DeleteTemplate deletes a workflow template
func (d TinkDB) DeleteTemplate(ctx context.Context, name string) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	UPDATE template
	SET
		deleted_at = NOW()
	WHERE
		id = $1;
	`, name)
	if err != nil {
		return errors.Wrap(err, "UPDATE")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// ListTemplates returns all saved templates
func (d TinkDB) ListTemplates(fn func(id, n string, in, del *timestamp.Timestamp) error) error {
	rows, err := d.instance.Query(`
	SELECT id, name, created_at, updated_at
	FROM template
	WHERE
		deleted_at IS NULL;
	`)

	if err != nil {
		return err
	}

	defer rows.Close()
	var (
		id        string
		name      string
		createdAt time.Time
		updatedAt time.Time
	)

	for rows.Next() {
		err = rows.Scan(&id, &name, &createdAt, &updatedAt)
		if err != nil {
			err = errors.Wrap(err, "SELECT")
			logger.Error(err)
			return err
		}

		tCr, _ := ptypes.TimestampProto(createdAt)
		tUp, _ := ptypes.TimestampProto(updatedAt)
		err = fn(id, name, tCr, tUp)
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

// UpdateTemplate update a given template
func (d TinkDB) UpdateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	if data == "" && name != "" {
		_, err = tx.Exec(`
		UPDATE template
		SET
			updated_at = NOW(), name = $2
		WHERE
			id = $1;`, id, name)
	} else if data != "" && name == "" {
		_, err = tx.Exec(`
		UPDATE template
		SET
			updated_at = NOW(), data = $2
		WHERE
			id = $1;`, id, data)
	} else {
		_, err = tx.Exec(`
		UPDATE template
		SET
			updated_at = NOW(), name = $2, data = $3
		WHERE
			id = $1;
		`, id, name, data)
	}

	if err != nil {
		return errors.Wrap(err, "UPDATE")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}
