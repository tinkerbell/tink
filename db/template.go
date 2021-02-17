package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	wflow "github.com/tinkerbell/tink/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateTemplate creates a new workflow template
func (d TinkDB) CreateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {
	_, err := wflow.Parse([]byte(data))
	if err != nil {
		return err
	}

	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
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

// GetTemplate returns template which is not deleted
func (d TinkDB) GetTemplate(ctx context.Context, fields map[string]string, deleted bool) (string, string, string, error) {
	getCondition, err := buildGetCondition(fields)
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to get template")
	}

	var query string
	if !deleted {
		query = `
	SELECT id, name, data
	FROM template
	WHERE
		` + getCondition + ` AND
		deleted_at IS NULL
	`
	} else {
		query = `
	SELECT id, name, data
	FROM template
	WHERE
		` + getCondition + `
	`
	}

	row := d.instance.QueryRowContext(ctx, query)
	id := []byte{}
	name := []byte{}
	data := []byte{}
	err = row.Scan(&id, &name, &data)
	if err == nil {
		return string(id), string(name), string(data), nil
	}
	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT")
		d.logger.Error(err)
	}
	return "", "", "", err
}

// DeleteTemplate deletes a workflow template by id
func (d TinkDB) DeleteTemplate(ctx context.Context, id string) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	res, err := tx.Exec(`
	UPDATE template
	SET
		deleted_at = NOW()
	WHERE
		id = $1;
	`, id)
	if err != nil {
		return errors.Wrap(err, "UPDATE")
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

// ListTemplates returns all saved templates
func (d TinkDB) ListTemplates(filter string, fn func(id, n string, in, del *timestamp.Timestamp) error) error {
	rows, err := d.instance.Query(`
	SELECT id, name, created_at, updated_at
	FROM template
	WHERE
		name ILIKE $1
	AND
		deleted_at IS NULL;
	`, filter)

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
			d.logger.Error(err)
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

// ListTemplateRevisions returns revisions saved for a given template
func (d TinkDB) ListTemplateRevisions(id string, fn func(id string, revision int, tCr *timestamp.Timestamp) error) error {
	rows, err := d.instance.Query(`
		SELECT revision, created_at
		FROM template_revisions
		WHERE
			template_id ILIKE $1
		AND
			deleted_at IS NULL;
	`, id)

	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		revision  int
		createdAt time.Time
	)
	for rows.Next() {
		err = rows.Scan(&revision, &createdAt)
		if err != nil {
			err = errors.Wrap(err, "SELECT")
			d.logger.Error(err)
			return err
		}

		tCr, _ := ptypes.TimestampProto(createdAt)
		err = fn(id, revision, tCr)
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
