package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

	revision, err := writeTemplateRevision(ctx, d.instance, tx, id, data)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	_, err = tx.Exec(`
		INSERT INTO
			template (created_at, updated_at, name, revision, id)
		VALUES
			($1, $1, $2, $3, $4)
		ON CONFLICT (id)
		DO
		UPDATE SET
			(updated_at, deleted_at, name, revision) = ($1, NULL, $2, $3);
	`, time.Now(), name, revision, id)
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
	SELECT id, name, revision
	FROM template
	WHERE
		` + getCondition + ` AND
		deleted_at IS NULL
	`
	} else {
		query = `
	SELECT id, name, revision
	FROM template
	WHERE
		` + getCondition + `
	`
	}

	row := d.instance.QueryRowContext(ctx, query)
	id := []byte{}
	name := []byte{}
	revision := 0
	err = row.Scan(&id, &name, &revision)
	if err == nil {
		data, err := getTemplate(ctx, d.instance, string(id), revision)
		if err == nil {
			return string(id), string(name), data, nil
		}
		if err != sql.ErrNoRows {
			d.logger.Error(err)
			return "", "", "", errors.Wrap(err, "SELECT")
		}
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

	_, err = tx.Exec(`
	UPDATE template_revisions
	SET
		deleted_at = NOW()
	WHERE
		template_id = $1;
	`, id)
	if err != nil {
		return errors.Wrap(err, "UPDATE")
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

	revision, err := writeTemplateRevision(ctx, d.instance, tx, id, data)
	if err != nil {
		return errors.Wrap(err, "UPDATE")
	}
	_, err = tx.Exec(`
		UPDATE template
		SET
			updated_at = NOW(), name = $2, revision = $3
		WHERE
			id = $1;
		`, id, name, revision)
	if err != nil {
		return errors.Wrap(err, "UPDATE")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// ListRevisionsByTemplateID returns revisions saved for a given template
func (d TinkDB) ListRevisionsByTemplateID(id string, fn func(revision int, tCr *timestamp.Timestamp) error) error {
	query := `SELECT revision, created_at FROM template_revisions
		WHERE template_id='` + id + `' AND deleted_at IS NULL;`
	rows, err := d.instance.Query(query)
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
		err = fn(revision, tCr)
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

func writeTemplateRevision(ctx context.Context, db *sql.DB, tx *sql.Tx, templateID uuid.UUID, data string) (int, error) {
	revision, err := getLatestRevision(ctx, db, templateID)
	if err != nil {
		return revision, err
	}

	_, err = tx.Exec(`
		INSERT INTO template_revisions (template_id, revision, data)
		VALUES ($1, $2, $3)`, templateID, revision+1, data)
	if err != nil {
		return revision, errors.Wrap(err, "INSERT")
	}
	return revision + 1, nil
}

func getLatestRevision(ctx context.Context, db *sql.DB, templateID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(revision) FROM template_revisions
		WHERE template_id='` + templateID.String() + `' AND deleted_at IS NULL;
	`
	row := db.QueryRowContext(ctx, query)
	var revision int
	err := row.Scan(&revision)
	return revision, err
}

func getTemplate(ctx context.Context, db *sql.DB, id string, r int) (string, error) {
	query := `SELECT data FROM template_revisions
			WHERE template_id='` + id + `' AND revision=` + strconv.Itoa(r) +
		` AND deleted_at is NULL`

	row := db.QueryRowContext(ctx, query)
	data := []byte{}
	err := row.Scan(&data)
	if err == nil {
		return string(data), nil
	}
	if err != sql.ErrNoRows {
		return "", errors.Wrap(err, "SELECT")
	}
	return "", err
}
