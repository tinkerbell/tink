package executor

import (
	"context"
	"database/sql"
	"fmt"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/packethost/rover/db"
	pb "github.com/packethost/rover/protos/rover"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	yaml "gopkg.in/yaml.v2"
)

var (
	workflowcontexts = map[string]*pb.WorkflowContext{}
	workflowactions  = map[string]*pb.WorkflowActionList{}
	workers          = map[string][]string{}
)

// LoadWorkflow loads workflow in memory and polulates required constructs
func LoadWorkflow(id, data string) error {
	var wf Workflow
	err := yaml.Unmarshal([]byte(data), &wf)
	if err != nil {
		return err
	}
	workflowcontexts[id] = &pb.WorkflowContext{WorkflowId: id}
	updateWorkflowActions(id, wf.Tasks)
	fmt.Println(workers)
	fmt.Println(*(workflowcontexts[id]))
	fmt.Println(*(workflowactions[id]))

	// ingest()
	return nil
}

func updateWorkflowActions(id string, tasks []Task) {
	list := []*pb.WorkflowAction{}
	for _, task := range tasks {
		for _, action := range task.Actions {
			list = append(list, &pb.WorkflowAction{
				WorkerId: task.Worker,
				TaskName: task.Name,
				Name:     action.Name,
				Image:    action.Image,
			})

			wfs := workers[task.Worker]
			add := true
			for _, wf := range wfs {
				if id == wf {
					add = false
					break
				}
			}
			if add {
				workers[task.Worker] = append(wfs, id)
			}
		}
	}
	workflowactions[id] = &pb.WorkflowActionList{ActionList: list}
}

// GetWorkflowContexts implements rover.GetWorkflowContexts
func GetWorkflowContexts(context context.Context, req *pb.WorkflowContextRequest) (*pb.WorkflowContextList, error) {
	if len(req.WorkerId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "worker_id is invalid")
	}
	wfs, ok := workers[req.WorkerId]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Worker not found for any workflows")
	}

	wfContexts := []*pb.WorkflowContext{}

	for _, wf := range wfs {
		wfContext, ok := workflowcontexts[wf]
		if !ok {
			return nil, status.Errorf(codes.Aborted, "Invalid workflow %s found for worker %s", wf, req.WorkerId)
		}
		wfContexts = append(wfContexts, wfContext)
	}

	return &pb.WorkflowContextList{
		WorkflowContexts: wfContexts,
	}, nil
}

// GetWorkflowActions implements rover.GetWorkflowActions
func GetWorkflowActions(context context.Context, req *pb.WorkflowActionsRequest) (*pb.WorkflowActionList, error) {
	wfID := req.GetWorkflowId()
	if len(wfID) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	actions, ok := workflowactions[wfID]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	return actions, nil
}

// ReportActionStatus implements rover.ReportActionStatus
func ReportActionStatus(context context.Context, req *pb.WorkflowActionStatus, sdb *sql.DB) (*empty.Empty, error) {
	wfID := req.GetWorkflowId()
	if len(wfID) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	if len(req.GetTaskName()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "task_name is invalid")
	}
	if len(req.GetActionName()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "action_name is invalid")
	}
	fmt.Printf("Received action status: %s\n", req)
	wfContext, ok := workflowcontexts[wfID]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Workflow context not found for workflow %s", wfID)
	}
	wfActions, ok := workflowactions[wfID]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Workflow actions not found for workflow %s", wfID)
	}

	// We need bunch of checks here considering
	// Considering concurrency and network latencies & accuracy for proceeding of WF
	actionIndex := wfContext.GetCurrentActionIndex()
	if req.GetActionStatus() == pb.ActionState_ACTION_IN_PROGRESS {
		if wfContext.GetCurrentAction() != "" {
			actionIndex = actionIndex + 1
		}
	}
	action := wfActions.ActionList[actionIndex]
	if action.GetTaskName() != req.GetTaskName() {
		return nil, status.Errorf(codes.FailedPrecondition, "Reported task name not matching in actions info")
	}
	if action.GetName() != req.GetActionName() {
		return nil, status.Errorf(codes.FailedPrecondition, "Reported action name not matching in actions info")
	}
	wfContext.CurrentWorker = action.GetWorkerId()
	wfContext.CurrentTask = req.GetTaskName()
	wfContext.CurrentAction = req.GetActionName()
	wfContext.CurrentActionState = req.GetActionStatus()
	wfContext.CurrentActionIndex = actionIndex
	err := db.UpdateWorkflowStateTable(context, sdb, wfContext)
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("Failed to update the workflow_state table. Error : %s", err)
	}
	err = db.InsertIntoWorkflowEventTable(context, sdb, req)
	if err != nil {
		return &empty.Empty{}, fmt.Errorf("Failed to update the workflow_event table. Error : %s", err)
	}
	fmt.Printf("Current context %s\n", wfContext)
	return &empty.Empty{}, nil
}
