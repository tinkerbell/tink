package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
)

// Workflow represents a workflow instance in database
type Workflow struct {
	State                int32
	ID, Target, Template string
	CreatedAt, UpdatedAt *timestamp.Timestamp
}

// CreateWorkflow creates a new workflow
func CreateWorkflow(ctx context.Context, db *sql.DB, wf Workflow) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	INSERT INTO
		workflow (created_at, updated_at, template, target, state, id)
	VALUES
		($1, $1, $2, $3, $4, $5)
	ON CONFLICT (id)
	DO
	UPDATE SET
		(updated_at, deleted_at, template, target, state) = ($1, NULL, $2, $3, $4);
	`, time.Now(), wf.Template, wf.Target, wf.State, wf.ID)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// GetWorkflow returns a workflow
func GetWorkflow(ctx context.Context, db *sql.DB, id string) (Workflow, error) {
	query := `
	SELECT template, target, state
	FROM workflow
	WHERE
		id = $1
	AND
		deleted_at IS NULL;
	`
	row := db.QueryRowContext(ctx, query, id)
	var tmp, tar string
	var state int32
	err := row.Scan(&tmp, &tar, &state)
	if err == nil {
		return Workflow{ID: id, Template: tmp, Target: tar, State: state}, nil
	}

	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT")
		logger.Error(err)
	} else {
		err = nil
	}

	return Workflow{}, nil
}

// DeleteWorkflow deletes a workflow
func DeleteWorkflow(ctx context.Context, db *sql.DB, id string, state int32) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	UPDATE workflow
	SET
		deleted_at = NOW()
	WHERE
		id = $1
	AND
		state != $2;
	`, id, state)
	if err != nil {
		return errors.Wrap(err, "UPDATE")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// ListWorkflows returns all workflows
func ListWorkflows(db *sql.DB, fn func(wf Workflow) error) error {
	rows, err := db.Query(`
	SELECT id, template, target, state, created_at, updated_at
	FROM workflow
	WHERE
		deleted_at IS NULL;
	`)

	if err != nil {
		return err
	}

	defer rows.Close()
	var (
		state        int32
		id, tmp, tar string
		crAt, upAt   time.Time
	)

	for rows.Next() {
		err = rows.Scan(&id, &tmp, &tar, &state, &crAt, &upAt)
		if err != nil {
			err = errors.Wrap(err, "SELECT")
			logger.Error(err)
			return err
		}

		wf := Workflow{
			ID:       id,
			Template: tmp,
			Target:   tar,
			State:    state,
		}
		wf.CreatedAt, _ = ptypes.TimestampProto(crAt)
		wf.UpdatedAt, _ = ptypes.TimestampProto(upAt)
		err = fn(wf)
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

// UpdateWorkflow updates a given workflow
func UpdateWorkflow(ctx context.Context, db *sql.DB, wf Workflow, state int32) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	if wf.Target == "" && wf.Template != "" {
		_, err = tx.Exec(`
		UPDATE workflow
		SET
			updated_at = NOW(), template = $2
		WHERE
			id = $1
		AND
			state != $3;`, wf.ID, wf.Template, state)
	} else if wf.Target != "" && wf.Template == "" {
		_, err = tx.Exec(`
		UPDATE workflow
		SET
			updated_at = NOW(), target = $2
		WHERE
			id = $1
		AND
			state != $3;`, wf.ID, wf.Target, state)
	} else {
		_, err = tx.Exec(`
		UPDATE workflow
		SET
			updated_at = NOW(), template = $2, target = $3
		WHERE
			id = $1
		AND
			state != $4;`, wf.ID, wf.Template, wf.Target, state)
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
