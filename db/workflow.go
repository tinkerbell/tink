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
	pb "github.com/packethost/rover/protos/rover"
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
		GlobalTimeout int    `yaml:"global_timeout"`
		Tasks         []Task `yaml:"tasks"`
	}

	// Task represents a task to be performed in a worflow
	Task struct {
		Name       string   `yaml:"name"`
		WorkerAddr string   `yaml:"worker"`
		Actions    []Action `yaml:"actions"`
	}

	// Action is the basic executional unit for a workflow
	Action struct {
		Name      string `yaml:"name"`
		Image     string `yaml:"image"`
		Timeout   int64  `yaml:"timeout"`
		Command   string `yaml:"command"`
		OnTimeout string `yaml:"on-timeout"`
		OnFailure string `yaml:"on-failure"`
	}
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
        SELECT id
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
		return getWorkerIDbyMac(ctx, db, addr)
	}
}

func insertIntoWfWorkerTable(ctx context.Context, db *sql.DB, wfID uuid.UUID, workerID uuid.UUID) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}
	_, err = tx.Exec(`
	INSERT INTO
		workflow_worker_map (workflow_id, worker_id)
	VALUES
		($1, $2);
	`, wfID, workerID)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow_worker_map")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

func validateUniqueTaskAndActionName(tasks []Task) error {
	taskNameMap := make(map[string]struct{})
	for _, task := range tasks {
		_, ok := taskNameMap[task.Name]
		if ok {
			return fmt.Errorf("Provided template has duplicate task name \"%s\"", task.Name)
		} else {
			taskNameMap[task.Name] = struct{}{}
			actionNameMap := make(map[string]struct{})
			for _, action := range task.Actions {
				_, ok := actionNameMap[action.Name]
				if ok {
					return fmt.Errorf("Provided template has duplicate action name \"%s\" in task \"%s\"", action.Name, task.Name)
				} else {
					actionNameMap[action.Name] = struct{}{}
				}
			}

		}
	}
	return nil
}

// Insert actions in the workflow_state table
func InsertActionList(ctx context.Context, db *sql.DB, yamlData string, id uuid.UUID) error {
	wfymldata, err := parseYaml([]byte(yamlData))
	if err != nil {
		return err
	}
	err = validateUniqueTaskAndActionName(wfymldata.Tasks)
	if err != nil {
		return errors.Wrap(err, "Invalid Template")
	}
	var actionList []pb.WorkflowAction
	var uniqueWorkerID uuid.UUID
	for _, task := range wfymldata.Tasks {
		workerID, err := getWorkerID(ctx, db, task.WorkerAddr)
		if err != nil {
			return err
		} else if workerID == "" {
			return fmt.Errorf("Target mentioned with refernece %s not found", task.WorkerAddr)
		}
		workerUID, err := uuid.FromString(workerID)
		if err != nil {
			return err
		}
		if uniqueWorkerID != workerUID {
			insertIntoWfWorkerTable(ctx, db, id, workerUID)
			uniqueWorkerID = workerUID
		}
		for _, ac := range task.Actions {
			action := pb.WorkflowAction{
				TaskName:  task.Name,
				WorkerId:  workerUID.String(),
				Name:      ac.Name,
				Image:     ac.Image,
				Timeout:   ac.Timeout,
				Command:   ac.Command,
				OnTimeout: ac.OnTimeout,
				OnFailure: ac.OnFailure,
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
		workflow_state (workflow_id, action_list, current_action_index)
	VALUES
		($1, $2, $3)
	ON CONFLICT (workflow_id)
	DO
	UPDATE SET
		(workflow_id, action_list, current_action_index) = ($1, $2, $3);
	`, id, actionData, 0)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow_state")
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

func UpdateWorkflowStateTable(ctx context.Context, db *sql.DB, wfContext *pb.WorkflowContext) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	UPDATE workflow_state
	SET current_task_name = $2,
		current_action_name = $3,
		current_action_state = $4, 
		current_worker = $5, 
		current_action_index = $6
	WHERE
		workflow_id = $1;
	`, wfContext.WorkflowId, wfContext.CurrentTask, wfContext.CurrentAction, wfContext.CurrentActionState, wfContext.CurrentWorker, wfContext.CurrentActionIndex)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow_state")
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

func InsertIntoWorkflowEventTable(ctx context.Context, db *sql.DB, wfEvent *pb.WorkflowActionStatus) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	INSERT INTO
		workflow_event (workflow_id, task_name, action_name, execution_time, message, status)
	VALUES
		($1, $2, $3, $4, $5, $6);
	`, wfEvent.WorkflowId, wfEvent.TaskName, wfEvent.ActionName, wfEvent.Seconds, wfEvent.Message, wfEvent.ActionStatus)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow_state")
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}
