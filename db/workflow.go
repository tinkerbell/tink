package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
)

type (
	// Workflow holds details about the workflow to be executed
	WfYamlstruct struct {
		Version       string `yaml:"version"`
		Name          string `yaml:"name"`
		ID            string `yaml:"id"`
		WorkflowID    string `yaml:"work_id"`
		GlobalTimeout int    `yaml:"global_timeout"`
		Tasks         []Task `yaml:"tasks"`
		//logger      *zap.logger
		Status string
	}

	// Task represents a task to be performed in a worflow
	Task struct {
		Name      string `yaml:"name"`
		WorkeAddr string `yaml:"worker"`
		Actions   []Act  `yaml:"actions"`
		Onfailure string `yaml:"on-failure"`
		Ontimeout string `yaml:"on-timeout"`
	}

	// Act is the basic executional unit for a workflow
	Act struct {
		Name      string `yaml:"name"`
		Image     string `yaml:"image"`
		Timeout   int    `yaml:"timeout"`
		Ontimeout string `yaml:"on-timeout"`
		Onfailure string `yaml:"on-failure"`
	}
)

type Action struct {
	TaskName  string
	WorkerID  uuid.UUID
	Name      string
	Image     string
	Timeout   int
	Ontimeout string
	Onfailure string
}

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
		return errors.Wrap(err, "INSERT in to workflow")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

func parseYaml(ymlContent []byte) (*WfYamlstruct, error) {
	var workflow = WfYamlstruct{}
	err := yaml.Unmarshal(ymlContent, &workflow)
	if err != nil {
		return &WfYamlstruct{}, err
	}
	return &workflow, nil
}

func getWorkerIDbyMac(ctx context.Context, db *sql.DB, mac string) (string, error) {
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
	SELECT id
	FROM hardware
	WHERE
		deleted_at IS NULL
	AND
		data @> $1
	`

	return get(ctx, db, query, arg)
}

func getWorkerIDbyIP(ctx context.Context, db *sql.DB, ip string) (string, error) {
	arg := ip

	query := `
	SELECT id
	FROM hardware
	WHERE
		deleted_at IS NULL
	AND
		id = $1
	`
	return get(ctx, db, query, arg)
}

func getWorkerID(ctx context.Context, db *sql.DB, addr string) (string, error) {
	_, err := net.ParseMAC(addr)
	if err != nil {
		ip := net.ParseIP(addr)
		if ip == nil || ip.To4() == nil {
			return "", fmt.Errorf("invalid worker address: %s", addr)
		} else {
			return getWorkerIDbyIP(ctx, db, addr)

		}
	} else {
		fmt.Println("Getting the worker ID by MAC")
		return getWorkerIDbyMac(ctx, db, addr)
	}
}

func insertIntoWfWorkerTable(ctx context.Context, db *sql.DB, wfId uuid.UUID, workerId uuid.UUID) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}
	_, err = tx.Exec(`
	INSERT INTO
		wfworker (wfid, worker)
	VALUES
		($1, $2);
	`, wfId, workerId)
	if err != nil {
		return errors.Wrap(err, "INSERT in to wfworker")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

func InsertActionList(ctx context.Context, db *sql.DB, yamlData string, id uuid.UUID) error {
	wfymldata, err := parseYaml([]byte(yamlData))
	if err != nil {
		return err
	}
	var actionList []Action
	var uniqueWorkerID uuid.UUID
	for _, task := range wfymldata.Tasks {
		workerID, err := getWorkerID(ctx, db, task.WorkeAddr)
		if err != nil {
			return err
		}
		fmt.Println("Worker ID", workerID)
		workerUID, err := uuid.FromString(workerID)
		if err != nil {
			return err
		}
		fmt.Println("Worker UID", workerUID)
		if uniqueWorkerID != workerUID {
			insertIntoWfWorkerTable(ctx, db, id, workerUID)
			uniqueWorkerID = workerUID
		}
		for _, ac := range task.Actions {
			action := Action{
				TaskName:  task.Name,
				WorkerID:  workerUID,
				Name:      ac.Name,
				Image:     ac.Image,
				Timeout:   ac.Timeout,
				Ontimeout: ac.Ontimeout,
				Onfailure: ac.Onfailure,
			}
			actionList = append(actionList, action)
		}
	}
	actionData, err := json.Marshal(actionList)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	INSERT INTO
		wfstate (wfid, actionList, currentActionIndex)
	VALUES
		($1, $2, $3)
	ON CONFLICT (wfid)
	DO
	UPDATE SET
		(wfid, actionList, currentActionIndex) = ($1, $2, $3);
	`, id, actionData, 0)
	if err != nil {
		return errors.Wrap(err, "INSERT in to wfstate")
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
