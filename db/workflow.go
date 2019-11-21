package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/packethost/rover/protos/rover"
	workflowpb "github.com/packethost/rover/protos/workflow"
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
func CreateWorkflow(ctx context.Context, db *sql.DB, wf Workflow, data string, id uuid.UUID) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	err = insertActionList(ctx, db, data, id, tx)
	if err != nil {
		return errors.Wrap(err, "Failed to insert in workflow_state")

	}
	err = insertInWorkflow(ctx, db, wf, tx)
	if err != nil {
		return errors.Wrap(err, "Failed to workflow")

	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

func insertInWorkflow(ctx context.Context, db *sql.DB, wf Workflow, tx *sql.Tx) error {
	_, err := tx.Exec(`
	INSERT INTO
		workflow (created_at, updated_at, template, target, id)
	VALUES
		($1, $1, $2, $3, $4)
	ON CONFLICT (id)
	DO
	UPDATE SET
		(updated_at, deleted_at, template, target) = ($1, NULL, $2, $3);
	`, time.Now(), wf.Template, wf.Target, wf.ID)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow")
	}
	return nil
}

func insertIntoWfWorkerTable(ctx context.Context, db *sql.DB, wfID uuid.UUID, workerID uuid.UUID, tx *sql.Tx) error {
	// TODO This command is not 100% reliable for concurrent write operations
	_, err := tx.Exec(`
	INSERT INTO
		workflow_worker_map (workflow_id, worker_id)
	SELECT $1, $2
	WHERE 
		NOT EXISTS (
			SELECT workflow_id FROM workflow_worker_map WHERE workflow_id = $1 AND worker_id = $2
		);
	`, wfID, workerID)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow_worker_map")
	}
	return nil
}

// Insert actions in the workflow_state table
func insertActionList(ctx context.Context, db *sql.DB, yamlData string, id uuid.UUID, tx *sql.Tx) error {
	wfymldata, err := parseYaml([]byte(yamlData))
	if err != nil {
		return err
	}
	err = validateTemplateValues(wfymldata.Tasks)
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
			err = insertIntoWfWorkerTable(ctx, db, id, workerUID, tx)
			if err != nil {
				return err
			}
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
	totalActions := int64(len(actionList))
	actionData, err := json.Marshal(actionList)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	INSERT INTO
		workflow_state (workflow_id, current_worker, current_task_name, current_action_name, current_action_state, action_list, current_action_index, total_number_of_actions)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (workflow_id)
	DO
	UPDATE SET
		(workflow_id, current_worker, current_task_name, current_action_name, current_action_state, action_list, current_action_index, total_number_of_actions) = ($1, $2, $3, $4, $5, $6, $7, $8);
	`, id, "", "", "", 0, actionData, 0, totalActions)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow_state")
	}
	return nil
}

func GetfromWfWorkflowTable(ctx context.Context, db *sql.DB, id string) ([]string, error) {
	rows, err := db.Query(`
	SELECT workflow_id
	FROM workflow_worker_map
	WHERE
		worker_id = $1;
	`, id)
	if err != nil {
		return nil, err
	}
	var wfID []string
	defer rows.Close()
	var workerId string

	for rows.Next() {
		err = rows.Scan(&workerId)
		if err != nil {
			err = errors.Wrap(err, "SELECT from worflow_worker_map")
			logger.Error(err)
			return nil, err
		}
		wfID = append(wfID, workerId)
	}
	err = rows.Err()
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return wfID, err
}

// GetWorkflow returns a workflow
func GetWorkflow(ctx context.Context, db *sql.DB, id string) (Workflow, error) {
	query := `
	SELECT template, target
	FROM workflow
	WHERE
		id = $1
	AND
		deleted_at IS NULL;
	`
	row := db.QueryRowContext(ctx, query, id)
	var tmp, tar string
	err := row.Scan(&tmp, &tar)
	if err == nil {
		return Workflow{ID: id, Template: tmp, Target: tar}, nil
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
	DELETE FROM workflow_worker_map
	WHERE
		workflow_id = $1;
	`, id)
	if err != nil {
		return errors.Wrap(err, "Delete Workflow Error")
	}

	_, err = tx.Exec(`
	DELETE FROM workflow_state
	WHERE
		workflow_id = $1;
	`, id)
	if err != nil {
		return errors.Wrap(err, "Delete Workflow Error")
	}

	_, err = tx.Exec(`
	UPDATE workflow
	SET
		deleted_at = NOW()
	WHERE
		id = $1;
	`, id)
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
	SELECT id, template, target, created_at, updated_at
	FROM workflow
	WHERE
		deleted_at IS NULL;
	`)

	if err != nil {
		return err
	}

	defer rows.Close()
	var (
		id, tmp, tar string
		crAt, upAt   time.Time
	)

	for rows.Next() {
		err = rows.Scan(&id, &tmp, &tar, &crAt, &upAt)
		if err != nil {
			err = errors.Wrap(err, "SELECT")
			logger.Error(err)
			return err
		}

		wf := Workflow{
			ID:       id,
			Template: tmp,
			Target:   tar,
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
			id = $1;
		`, wf.ID, wf.Template)
	} else if wf.Target != "" && wf.Template == "" {
		_, err = tx.Exec(`
		UPDATE workflow
		SET
			updated_at = NOW(), target = $2
		WHERE
			id = $1;
		`, wf.ID, wf.Target)
	} else {
		_, err = tx.Exec(`
		UPDATE workflow
		SET
			updated_at = NOW(), template = $2, target = $3
		WHERE
			id = $1;
		`, wf.ID, wf.Template, wf.Target)
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

func UpdateWorkflowState(ctx context.Context, db *sql.DB, wfContext *workflowpb.WorkflowContext) error {
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

func GetWorkflowContexts(ctx context.Context, db *sql.DB, wfId string) (*workflowpb.WorkflowContext, error) {
	query := `
	SELECT current_worker, current_task_name, current_action_name, current_action_index, current_action_state, total_number_of_actions
	FROM workflow_state
	WHERE
		workflow_id = $1;
	`
	row := db.QueryRowContext(ctx, query, wfId)
	var cw, ct, ca string
	var cai, tact int64
	var cas workflowpb.ActionState
	err := row.Scan(&cw, &ct, &ca, &cai, &cas, &tact)
	if err == nil {
		return &workflowpb.WorkflowContext{
			WorkflowId:           wfId,
			CurrentWorker:        cw,
			CurrentTask:          ct,
			CurrentAction:        ca,
			CurrentActionIndex:   cai,
			CurrentActionState:   cas,
			TotalNumberOfActions: tact}, nil
	}
	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT from worflow_state")
		logger.Error(err)
	} else {
		err = nil
	}
	return &workflowpb.WorkflowContext{}, nil
}

func GetWorkflowActions(ctx context.Context, db *sql.DB, wfId string) (*pb.WorkflowActionList, error) {
	query := `
	SELECT action_list
	FROM workflow_state
	WHERE
		workflow_id = $1;
	`
	row := db.QueryRowContext(ctx, query, wfId)
	var actionList string
	err := row.Scan(&actionList)
	if err == nil {
		actions := []*pb.WorkflowAction{}
		err = json.Unmarshal([]byte(actionList), &actions)
		return &pb.WorkflowActionList{
			ActionList: actions}, nil
	}
	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT from worflow_state")
		logger.Error(err)
	} else {
		err = nil
	}
	return &pb.WorkflowActionList{}, nil
}

func InsertIntoWorkflowEventTable(ctx context.Context, db *sql.DB, wfEvent *workflowpb.WorkflowActionStatus, time time.Time) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	// TODO "created_at" field should be set in worker and come in the request
	_, err = tx.Exec(`
	INSERT INTO
		workflow_event (workflow_id, worker_id, task_name, action_name, execution_time, message, status, created_at)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8);
	`, wfEvent.WorkflowId, wfEvent.WorkerId, wfEvent.TaskName, wfEvent.ActionName, wfEvent.Seconds, wfEvent.Message, wfEvent.ActionStatus, time)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow_event")
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// ShowWorkflowEvents returns all workflows
func ShowWorkflowEvents(db *sql.DB, wfId string, fn func(wfs workflowpb.WorkflowActionStatus) error) error {
	rows, err := db.Query(`
       SELECT worker_id, task_name, action_name, execution_time, message, status, created_at
	   FROM workflow_event
	   WHERE 
			   workflow_id = $1
		ORDER BY 
				created_at ASC;
	   `, wfId)

	if err != nil {
		return err
	}

	defer rows.Close()
	var (
		status                int32
		secs                  int64
		id, tName, aName, msg string
		evTime                time.Time
	)

	for rows.Next() {
		err = rows.Scan(&id, &tName, &aName, &secs, &msg, &status, &evTime)
		if err != nil {
			err = errors.Wrap(err, "SELECT")
			logger.Error(err)
			return err
		}
		createdAt, _ := ptypes.TimestampProto(evTime)
		wfs := workflowpb.WorkflowActionStatus{
			WorkerId:     id,
			TaskName:     tName,
			ActionName:   aName,
			Seconds:      secs,
			Message:      msg,
			ActionStatus: workflowpb.ActionState(status),
			CreatedAt:    createdAt,
		}
		err = fn(wfs)
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

func isValidLength(name string) error {
	if len(name) > 200 {
		return fmt.Errorf("Task/Action Name %s in the Temlate as more than 200 characters", name)
	}
	return nil
}

func isValidImageName(name string) error {
	_, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func validateTemplateValues(tasks []Task) error {
	taskNameMap := make(map[string]struct{})
	for _, task := range tasks {
		err := isValidLength(task.Name)
		if err != nil {
			return err
		}
		_, ok := taskNameMap[task.Name]
		if ok {
			return fmt.Errorf("Provided template has duplicate task name \"%s\"", task.Name)
		} else {
			taskNameMap[task.Name] = struct{}{}
			actionNameMap := make(map[string]struct{})
			for _, action := range task.Actions {
				err := isValidLength(action.Name)
				if err != nil {
					return err
				}
				err = isValidImageName(action.Image)
				if err != nil {
					return fmt.Errorf("Invalid Image name %s", action.Image)
				}

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
