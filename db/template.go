package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	tb "github.com/tinkerbell/tink/protos/template"
	wflow "github.com/tinkerbell/tink/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateTemplate creates a new workflow template.
func (d TinkDB) CreateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {
	_, err := wflow.Parse([]byte(data))
	if err != nil {
		return err
	}

	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}
	_, err = tx.ExecContext(ctx, `
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

// GetTemplate returns template which is not deleted.
func (d TinkDB) GetTemplate(ctx context.Context, fields map[string]string, deleted bool) (*tb.WorkflowTemplate, error) {
	getCondition, err := buildGetCondition(fields)
	if err != nil {
		return &tb.WorkflowTemplate{}, errors.Wrap(err, "failed to get template")
	}

	var query string
	if !deleted {
		query = `
	SELECT id, name, data, created_at, updated_at
	FROM template
	WHERE
		` + getCondition + ` AND
		deleted_at IS NULL
	`
	} else {
		query = `
	SELECT id, name, data, created_at, updated_at
	FROM template
	WHERE
		` + getCondition + `
	`
	}

	row := d.instance.QueryRowContext(ctx, query)
	var (
		id        string
		name      string
		data      string
		createdAt time.Time
		updatedAt time.Time
	)
	err = row.Scan(&id, &name, &data, &createdAt, &updatedAt)
	if err == nil {
		crAt := timestamppb.New(createdAt)
		upAt := timestamppb.New(updatedAt)
		return &tb.WorkflowTemplate{
			Id:        id,
			Name:      name,
			Data:      data,
			CreatedAt: crAt,
			UpdatedAt: upAt,
		}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		err = errors.Wrap(err, "SELECT")
		d.logger.Error(err)
	}
	return &tb.WorkflowTemplate{}, err
}

// DeleteTemplate deletes a workflow template by id.
func (d TinkDB) DeleteTemplate(ctx context.Context, id string) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	res, err := tx.ExecContext(ctx, `
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

// ListTemplates returns all saved templates.
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

		tCr := timestamppb.New(createdAt)
		tUp := timestamppb.New(updatedAt)
		err = fn(id, name, tCr, tUp)
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

// UpdateTemplate update a given template.
func (d TinkDB) UpdateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	switch {
	case data == "" && name != "":
		_, err = tx.ExecContext(ctx, `UPDATE template SET updated_at = NOW(), name = $2 WHERE id = $1;`, id, name)
	case data != "" && name == "":
		_, err = tx.ExecContext(ctx, `UPDATE template SET updated_at = NOW(), data = $2 WHERE id = $1;`, id, data)
	default:
		_, err = tx.ExecContext(ctx, `UPDATE template SET updated_at = NOW(), name = $2, data = $3 WHERE id = $1;`, id, name, data)
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
