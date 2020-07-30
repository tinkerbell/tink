package executor

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/tinkerbell/tink/db"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	workflowData = make(map[string]int)
)

// GetWorkflowContexts implements tinkerbell.GetWorkflowContexts
func GetWorkflowContexts(req *pb.WorkflowContextRequest, stream pb.WorkflowSvc_GetWorkflowContextsServer, sdb *sql.DB) error {
	if len(req.WorkerId) == 0 {
		return status.Errorf(codes.InvalidArgument, "worker_id is invalid")
	}
	wfs, _ := db.GetfromWfWorkflowTable(sdb, req.WorkerId)
	if wfs == nil {
		return status.Errorf(codes.InvalidArgument, "No workflow found for worker %s ", req.GetWorkerId())
	}

	for _, wf := range wfs {
		wfContext, err := db.GetWorkflowContexts(context.Background(), sdb, wf)
		if err != nil {
			return status.Errorf(codes.Aborted, "invalid workflow %s found for worker %s", wf, req.WorkerId)
		}
		if isApplicableToSend(context.Background(), wfContext, req.WorkerId, sdb) {
			stream.Send(wfContext)
		}
	}
	return nil
}

// GetWorkflowActions implements tinkerbell.GetWorkflowActions
func GetWorkflowActions(context context.Context, req *pb.WorkflowActionsRequest, sdb *sql.DB) (*pb.WorkflowActionList, error) {
	wfID := req.GetWorkflowId()
	if len(wfID) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	actions, err := db.GetWorkflowActions(context, sdb, wfID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	return actions, nil
}

// ReportActionStatus implements tinkerbell.ReportActionStatus
func ReportActionStatus(context context.Context, req *pb.WorkflowActionStatus, sdb *sql.DB) (*pb.Empty, error) {
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
	wfContext, err := db.GetWorkflowContexts(context, sdb, wfID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "workflow context not found for workflow %s", wfID)
	}
	wfActions, err := db.GetWorkflowActions(context, sdb, wfID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "workflow actions not found for workflow %s", wfID)
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
		return nil, status.Errorf(codes.FailedPrecondition, "reported task name not matching in actions info")
	}
	if action.GetName() != req.GetActionName() {
		return nil, status.Errorf(codes.FailedPrecondition, "reported action name not matching in actions info")
	}
	wfContext.CurrentWorker = action.GetWorkerId()
	wfContext.CurrentTask = req.GetTaskName()
	wfContext.CurrentAction = req.GetActionName()
	wfContext.CurrentActionState = req.GetActionStatus()
	wfContext.CurrentActionIndex = actionIndex
	err = db.UpdateWorkflowState(context, sdb, wfContext)
	if err != nil {
		return &pb.Empty{}, fmt.Errorf("failed to update the workflow_state table. Error : %s", err)
	}
	// TODO the below "time" would be a part of the request which is coming form worker.
	time := time.Now()
	err = db.InsertIntoWorkflowEventTable(context, sdb, req, time)
	if err != nil {
		return &pb.Empty{}, fmt.Errorf("failed to update the workflow_event table. Error : %s", err)
	}
	fmt.Printf("Current context %s\n", wfContext)
	return &pb.Empty{}, nil
}

// UpdateWorkflowData updates workflow ephemeral data
func UpdateWorkflowData(context context.Context, req *pb.UpdateWorkflowDataRequest, sdb *sql.DB) (*pb.Empty, error) {
	wfID := req.GetWorkflowID()
	if len(wfID) == 0 {
		return &pb.Empty{}, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	index, ok := workflowData[wfID]
	if ok {
		index = index + 1
	} else {
		index = 1
		workflowData[wfID] = index
	}
	err := db.InsertIntoWfDataTable(context, sdb, req)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Unknown, err.Error())
	}
	return &pb.Empty{}, nil
}

// GetWorkflowData returns ephemeral data for a particular workflow
func GetWorkflowData(context context.Context, req *pb.GetWorkflowDataRequest, sdb *sql.DB) (*pb.GetWorkflowDataResponse, error) {
	wfID := req.GetWorkflowID()
	if len(wfID) == 0 {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	data, err := db.GetfromWfDataTable(context, sdb, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Unknown, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow
func GetWorkflowMetadata(context context.Context, req *pb.GetWorkflowDataRequest, sdb *sql.DB) (*pb.GetWorkflowDataResponse, error) {
	data, err := db.GetWorkflowMetadata(context, sdb, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Unknown, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowDataVersion returns the latest version of data for a workflow
func GetWorkflowDataVersion(context context.Context, workflowID string, sdb *sql.DB) (*pb.GetWorkflowDataResponse, error) {
	version, err := db.GetWorkflowDataVersion(context, sdb, workflowID)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Version: version}, status.Errorf(codes.Unknown, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Version: version}, nil
}

// The below function check whether a particular workflow context is applicable or needed to
// be send to a worker based on the state of the current action and the targeted workerID.
func isApplicableToSend(context context.Context, wfContext *pb.WorkflowContext, workerID string, sdb *sql.DB) bool {
	if wfContext.GetCurrentActionState() == pb.ActionState_ACTION_FAILED ||
		wfContext.GetCurrentActionState() == pb.ActionState_ACTION_TIMEOUT {
		return false
	}
	actions, err := GetWorkflowActions(context, &pb.WorkflowActionsRequest{WorkflowId: wfContext.GetWorkflowId()}, sdb)
	if err != nil {
		return false
	}
	if wfContext.GetCurrentActionState() == pb.ActionState_ACTION_SUCCESS {
		if isLastAction(wfContext, actions) {
			return false
		}
		if wfContext.GetCurrentActionIndex() == 0 {
			if actions.ActionList[wfContext.GetCurrentActionIndex()+1].GetWorkerId() == workerID {
				log.Println("Send the workflow context ", wfContext.GetWorkflowId())
				return true
			}
		}
	} else {
		if actions.ActionList[wfContext.GetCurrentActionIndex()].GetWorkerId() == workerID {
			log.Println("Send the workflow context ", wfContext.GetWorkflowId())
			return true
		}
	}
	return false
}

func isLastAction(wfContext *pb.WorkflowContext, actions *pb.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}
