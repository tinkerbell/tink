package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
	wflow "github.com/tinkerbell/tink/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Workflow represents a workflow instance in database.
type Workflow struct {
	State                  int32
	ID, Hardware, Template string
	CreatedAt, UpdatedAt   *timestamp.Timestamp
}

var (
	defaultMaxVersions = 3
	maxVersions        = defaultMaxVersions // maximum number of workflow data versions to be kept in database
)

// CreateWorkflow creates a new workflow.
func (d TinkDB) CreateWorkflow(ctx context.Context, wf Workflow, data string, id uuid.UUID) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	err = insertActionList(ctx, d.instance, data, id, tx)
	if err != nil {
		return errors.Wrap(err, "failed to create workflow")
	}
	err = insertInWorkflow(ctx, d.instance, wf, tx)
	if err != nil {
		return errors.Wrap(err, "failed to create workflow")
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
		workflow (created_at, updated_at, template, devices, id)
	VALUES
		($1, $1, $2, $3, $4)
	ON CONFLICT (id)
	DO
	UPDATE SET
		(updated_at, deleted_at, template, devices) = ($1, NULL, $2, $3);
	`, time.Now(), wf.Template, wf.Hardware, wf.ID)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow")
	}
	return nil
}

func insertIntoWfWorkerTable(ctx context.Context, db *sql.DB, wfID uuid.UUID, workerID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec(`
	INSERT INTO
		workflow_worker_map (workflow_id, worker_id)
	VALUES
	        ($1, $2)
	ON CONFLICT (workflow_id, worker_id)
	DO NOTHING;
	`, wfID, workerID)
	if err != nil {
		return errors.Wrap(err, "INSERT in to workflow_worker_map")
	}
	return nil
}

// Insert actions in the workflow_state table.
func insertActionList(ctx context.Context, db *sql.DB, yamlData string, id uuid.UUID, tx *sql.Tx) error {
	wf, err := wflow.Parse([]byte(yamlData))
	if err != nil {
		return err
	}

	var actionList []*pb.WorkflowAction
	var uniqueWorkerID uuid.UUID
	for _, task := range wf.Tasks {
		taskEnvs := map[string]string{}
		taskVolumes := map[string]string{}
		for _, vol := range task.Volumes {
			v := strings.Split(vol, ":")
			taskVolumes[v[0]] = strings.Join(v[1:], ":")
		}
		for key, val := range task.Environment {
			taskEnvs[key] = val
		}

		workerID, err := getWorkerID(ctx, db, task.WorkerAddr)
		if err != nil {
			return errors.WithMessage(err, "unable to insert into action list")
		}
		workerUID, err := uuid.Parse(workerID)
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
			acenvs := map[string]string{}
			for key, val := range taskEnvs {
				acenvs[key] = val
			}
			for key, val := range ac.Environment {
				acenvs[key] = val
			}

			envs := []string{}
			for key, val := range acenvs {
				envs = append(envs, key+"="+val)
			}

			volumes := map[string]string{}
			for k, v := range taskVolumes {
				volumes[k] = v
			}

			for _, vol := range ac.Volumes {
				v := strings.Split(vol, ":")
				volumes[v[0]] = strings.Join(v[1:], ":")
			}

			ac.Volumes = []string{}
			for k, v := range volumes {
				ac.Volumes = append(ac.Volumes, k+":"+v)
			}

			action := pb.WorkflowAction{
				TaskName:    task.Name,
				WorkerId:    workerUID.String(),
				Name:        ac.Name,
				Image:       ac.Image,
				Timeout:     ac.Timeout,
				Command:     ac.Command,
				OnTimeout:   ac.OnTimeout,
				OnFailure:   ac.OnFailure,
				Environment: envs,
				Volumes:     ac.Volumes,
				Pid:         ac.Pid,
			}
			actionList = append(actionList, &action)
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

// InsertIntoWfDataTable : Insert ephemeral data in workflow_data table.
func (d TinkDB) InsertIntoWfDataTable(ctx context.Context, req *pb.UpdateWorkflowDataRequest) error {
	version, err := getLatestVersionWfData(ctx, d.instance, req.GetWorkflowId())
	if err != nil {
		return err
	}

	// increment version
	version = version + 1
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	_, err = tx.Exec(`
	INSERT INTO
		workflow_data (workflow_id, version, metadata, data)
	VALUES
		($1, $2, $3, $4);
	`, req.GetWorkflowId(), version, string(req.GetMetadata()), string(req.GetData()))
	if err != nil {
		return errors.Wrap(err, "INSERT Into workflow_data")
	}

	if version > int32(maxVersions) {
		cleanVersion := version - int32(maxVersions)
		_, err = tx.Exec(`
		UPDATE workflow_data
		SET
			data = NULL
		WHERE
			workflow_id = $1 AND version = $2;
		`, req.GetWorkflowId(), cleanVersion)
		if err != nil {
			return errors.Wrap(err, "UPDATE")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "COMMIT")
	}
	return nil
}

// GetfromWfDataTable : Give you the ephemeral data from workflow_data table.
func (d TinkDB) GetfromWfDataTable(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
	version := req.GetVersion()
	if req.Version == 0 {
		v, err := getLatestVersionWfData(ctx, d.instance, req.GetWorkflowId())
		if err != nil {
			return []byte(""), err
		}
		version = v
	}
	query := `
	SELECT data
	FROM workflow_data
	WHERE
		workflow_id = $1 AND version = $2
	`
	row := d.instance.QueryRowContext(ctx, query, req.GetWorkflowId(), version)
	buf := []byte{}
	err := row.Scan(&buf)
	if err == nil {
		return []byte(buf), nil
	}

	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT")
		d.logger.Error(err)
	}

	return []byte{}, nil
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow.
func (d TinkDB) GetWorkflowMetadata(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
	version := req.GetVersion()
	if req.Version == 0 {
		v, err := getLatestVersionWfData(ctx, d.instance, req.GetWorkflowId())
		if err != nil {
			return []byte(""), err
		}
		version = v
	}
	query := `
	SELECT metadata
	FROM workflow_data
	WHERE
		workflow_id = $1 AND version = $2
	`
	row := d.instance.QueryRowContext(ctx, query, req.GetWorkflowId(), version)
	buf := []byte{}
	err := row.Scan(&buf)
	if err == nil {
		return []byte(buf), nil
	}

	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT from workflow_data")
		d.logger.Error(err)
	}

	return []byte{}, nil
}

// GetWorkflowDataVersion returns the latest version of data for a workflow.
func (d TinkDB) GetWorkflowDataVersion(ctx context.Context, workflowID string) (int32, error) {
	return getLatestVersionWfData(ctx, d.instance, workflowID)
}

// GetWorkflowsForWorker : returns the list of workflows for a particular worker.
func (d TinkDB) GetWorkflowsForWorker(id string) ([]string, error) {
	rows, err := d.instance.Query(`
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
	var workerID string

	for rows.Next() {
		err = rows.Scan(&workerID)
		if err != nil {
			err = errors.Wrap(err, "SELECT from worflow_worker_map")
			d.logger.Error(err)
			return nil, err
		}
		wfID = append(wfID, workerID)
	}
	err = rows.Err()
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return wfID, err
}

// GetWorkflow returns a workflow.
func (d TinkDB) GetWorkflow(ctx context.Context, id string) (Workflow, error) {
	query := `
	SELECT template, devices, created_at, updated_at
	FROM workflow
	WHERE
		id = $1
	AND
		deleted_at IS NULL;
	`
	row := d.instance.QueryRowContext(ctx, query, id)
	var (
		tmp, tar   string
		crAt, upAt time.Time
	)
	err := row.Scan(&tmp, &tar, &crAt, &upAt)
	if err == nil {
		createdAt := timestamppb.New(crAt)
		updatedAt := timestamppb.New(upAt)
		return Workflow{
			ID:        id,
			Template:  tmp,
			Hardware:  tar,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, nil
	}
	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT")
		d.logger.Error(err)
		return Workflow{}, err
	}

	return Workflow{}, errors.New("Workflow with id " + id + " does not exist")
}

// DeleteWorkflow deletes a workflow.
func (d TinkDB) DeleteWorkflow(ctx context.Context, id string, state int32) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
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

	res, err := tx.Exec(`
	UPDATE workflow
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

// ListWorkflows returns all workflows.
func (d TinkDB) ListWorkflows(fn func(wf Workflow) error) error {
	rows, err := d.instance.Query(`
	SELECT id, template, devices, created_at, updated_at
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
			d.logger.Error(err)
			return err
		}

		wf := Workflow{
			ID:       id,
			Template: tmp,
			Hardware: tar,
		}
		wf.CreatedAt = timestamppb.New(crAt)
		wf.UpdatedAt = timestamppb.New(upAt)
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

// UpdateWorkflow updates a given workflow.
func (d TinkDB) UpdateWorkflow(ctx context.Context, wf Workflow, state int32) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.Wrap(err, "BEGIN transaction")
	}

	if wf.Hardware == "" && wf.Template != "" {
		_, err = tx.Exec(`
		UPDATE workflow
		SET
			updated_at = NOW(), template = $2
		WHERE
			id = $1;
		`, wf.ID, wf.Template)
	} else if wf.Hardware != "" && wf.Template == "" {
		_, err = tx.Exec(`
		UPDATE workflow
		SET
			updated_at = NOW(), devices = $2
		WHERE
			id = $1;
		`, wf.ID, wf.Hardware)
	} else {
		_, err = tx.Exec(`
		UPDATE workflow
		SET
			updated_at = NOW(), template = $2, devices = $3
		WHERE
			id = $1;
		`, wf.ID, wf.Template, wf.Hardware)
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

// UpdateWorkflowState : update the current workflow state.
func (d TinkDB) UpdateWorkflowState(ctx context.Context, wfContext *pb.WorkflowContext) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
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

// GetWorkflowContexts : gives you the current workflow context.
func (d TinkDB) GetWorkflowContexts(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
	query := `
	SELECT current_worker, current_task_name, current_action_name, current_action_index, current_action_state, total_number_of_actions
	FROM workflow_state
	WHERE
		workflow_id = $1;
	`
	row := d.instance.QueryRowContext(ctx, query, wfID)
	var cw, ct, ca string
	var cai, tact int64
	var cas pb.State
	err := row.Scan(&cw, &ct, &ca, &cai, &cas, &tact)
	if err == nil {
		return &pb.WorkflowContext{
			WorkflowId:           wfID,
			CurrentWorker:        cw,
			CurrentTask:          ct,
			CurrentAction:        ca,
			CurrentActionIndex:   cai,
			CurrentActionState:   cas,
			TotalNumberOfActions: tact,
		}, nil
	}
	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT from worflow_state")
		d.logger.Error(err)
		return &pb.WorkflowContext{}, err
	}
	return &pb.WorkflowContext{}, errors.New("Workflow with id " + wfID + " does not exist")
}

// GetWorkflowActions : gives you the action list of workflow.
func (d TinkDB) GetWorkflowActions(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
	query := `
	SELECT action_list
	FROM workflow_state
	WHERE
		workflow_id = $1;
	`
	row := d.instance.QueryRowContext(ctx, query, wfID)
	var actionList string
	err := row.Scan(&actionList)
	if err == nil {
		actions := []*pb.WorkflowAction{}
		if err := json.Unmarshal([]byte(actionList), &actions); err != nil {
			return nil, err
		}
		return &pb.WorkflowActionList{
			ActionList: actions,
		}, nil
	}
	if err != sql.ErrNoRows {
		err = errors.Wrap(err, "SELECT from worflow_state")
		d.logger.Error(err)
	}
	return &pb.WorkflowActionList{}, nil
}

// InsertIntoWorkflowEventTable : insert workflow event table.
func (d TinkDB) InsertIntoWorkflowEventTable(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
	tx, err := d.instance.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
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

// ShowWorkflowEvents returns all workflows.
func (d TinkDB) ShowWorkflowEvents(wfID string, fn func(wfs *pb.WorkflowActionStatus) error) error {
	rows, err := d.instance.Query(`
       SELECT worker_id, task_name, action_name, execution_time, message, status, created_at
	   FROM workflow_event
	   WHERE
			   workflow_id = $1
		ORDER BY
				created_at ASC;
	   `, wfID)
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
			d.logger.Error(err)
			return err
		}
		createdAt := timestamppb.New(evTime)
		wfs := &pb.WorkflowActionStatus{
			WorkerId:     id,
			TaskName:     tName,
			ActionName:   aName,
			Seconds:      secs,
			Message:      msg,
			ActionStatus: pb.State(status),
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

func getLatestVersionWfData(ctx context.Context, db *sql.DB, wfID string) (int32, error) {
	query := `
	SELECT COUNT(*)
	FROM workflow_data
	WHERE
		workflow_id = $1;
	`
	row := db.QueryRowContext(ctx, query, wfID)
	var version int32
	err := row.Scan(&version)
	if err != nil {
		return -1, err
	}
	return version, nil
}

func getWorkerIDbyMac(ctx context.Context, db *sql.DB, mac string) (string, error) {
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
	SELECT id
	FROM hardware
	WHERE
		deleted_at IS NULL
	AND
		data @> $1
	`

	id, err := get(ctx, db, query, arg)
	if errors.Cause(err) == sql.ErrNoRows {
		err = errors.WithMessage(errors.New(mac), "mac")
	}
	return id, err
}

func getWorkerIDbyIP(ctx context.Context, db *sql.DB, ip string) (string, error) {
	// update for instance (under metadata)
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

	id, err := get(ctx, db, query, instance, hardwareOrManagement)
	if errors.Cause(err) == sql.ErrNoRows {
		err = errors.WithMessage(errors.New(ip), "ip")
	}
	return id, err
}

func getWorkerID(ctx context.Context, db *sql.DB, addr string) (string, error) {
	parsedMAC, err := net.ParseMAC(addr)
	if err != nil {
		ip := net.ParseIP(addr)
		if ip == nil || ip.To4() == nil {
			return "", fmt.Errorf("invalid worker address: %s", addr)
		}
		id, err := getWorkerIDbyIP(ctx, db, addr)
		return id, errors.WithMessage(err, "no worker found")
	}
	id, err := getWorkerIDbyMac(ctx, db, parsedMAC.String())
	return id, errors.WithMessage(err, "no worker found")
}

func init() {
	val := os.Getenv("MAX_WORKFLOW_DATA_VERSIONS")
	if v, err := strconv.Atoi(val); err == nil {
		maxVersions = v
	}
}
