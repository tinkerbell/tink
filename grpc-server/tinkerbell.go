package grpcserver

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tinkerbell/tink/db"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var workflowData = make(map[string]int)

// GetWorkflowContexts implements tinkerbell.GetWorkflowContexts
func (s *server) GetWorkflowContexts(req *pb.WorkflowContextRequest, stream pb.WorkflowSvc_GetWorkflowContextsServer) error {
	wfs, err := getWorkflowsForWorker(s.db, req.WorkerId)
	if err != nil {
		return err
	}
	for _, wf := range wfs {
		wfContext, err := s.db.GetWorkflowContexts(context.Background(), wf)
		if err != nil {
			return status.Errorf(codes.Aborted, "invalid workflow %s found for worker %s", wf, req.WorkerId)
		}
		if isApplicableToSend(context.Background(), wfContext, req.WorkerId, s.db) {
			if err := stream.Send(wfContext); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetWorkflowContextList implements tinkerbell.GetWorkflowContextList
func (s *server) GetWorkflowContextList(context context.Context, req *pb.WorkflowContextRequest) (*pb.WorkflowContextList, error) {
	wfs, err := getWorkflowsForWorker(s.db, req.WorkerId)
	if err != nil {
		return nil, err
	}
	wfContexts := []*pb.WorkflowContext{}
	for _, wf := range wfs {
		wfContext, err := s.db.GetWorkflowContexts(context, wf)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, "Invalid workflow %s found for worker %s", wf, req.WorkerId)
		}
		wfContexts = append(wfContexts, wfContext)
	}
	return &pb.WorkflowContextList{
		WorkflowContexts: wfContexts,
	}, nil
}

// GetWorkflowActions implements tinkerbell.GetWorkflowActions
func (s *server) GetWorkflowActions(context context.Context, req *pb.WorkflowActionsRequest) (*pb.WorkflowActionList, error) {
	wfID := req.GetWorkflowId()
	if len(wfID) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	return getWorkflowActions(context, s.db, wfID)
}

// ReportActionStatus implements tinkerbell.ReportActionStatus
func (s *server) ReportActionStatus(context context.Context, req *pb.WorkflowActionStatus) (*pb.Empty, error) {
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
	wfContext, err := s.db.GetWorkflowContexts(context, wfID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "workflow context not found for workflow %s", wfID)
	}
	wfActions, err := s.db.GetWorkflowActions(context, wfID)
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
	err = s.db.UpdateWorkflowState(context, wfContext)
	if err != nil {
		return &pb.Empty{}, fmt.Errorf("failed to update the workflow_state table. Error : %s", err)
	}
	// TODO the below "time" would be a part of the request which is coming form worker.
	time := time.Now()
	err = s.db.InsertIntoWorkflowEventTable(context, req, time)
	if err != nil {
		return &pb.Empty{}, fmt.Errorf("failed to update the workflow_event table. Error : %s", err)
	}
	fmt.Printf("Current context %s\n", wfContext)
	return &pb.Empty{}, nil
}

// Update Workflow Ephemeral Data
func (s *server) UpdateWorkflowData(context context.Context, req *pb.UpdateWorkflowDataRequest) (*pb.Empty, error) {
	wfID := req.GetWorkflowID()
	if len(wfID) == 0 {
		return &pb.Empty{}, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	_, ok := workflowData[wfID]
	if !ok {
		workflowData[wfID] = 1
	}
	err := s.db.InsertIntoWfDataTable(context, req)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Unknown, err.Error())
	}
	return &pb.Empty{}, nil
}

// Get Workflow Ephemeral Data
func (s *server) GetWorkflowData(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	wfID := req.GetWorkflowID()
	if len(wfID) == 0 {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	data, err := s.db.GetfromWfDataTable(context, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Unknown, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow
func (s *server) GetWorkflowMetadata(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	data, err := s.db.GetWorkflowMetadata(context, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Unknown, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowDataVersion returns the latest version of data for a workflow
func (s *server) GetWorkflowDataVersion(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	version, err := s.db.GetWorkflowDataVersion(context, req.WorkflowID)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Version: version}, status.Errorf(codes.Unknown, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Version: version}, nil
}

func getWorkflowsForWorker(db db.Database, id string) ([]string, error) {
	if len(id) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "worker_id is invalid")
	}
	wfs, _ := db.GetWorkflowsForWorker(id)
	if wfs == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Worker not found for any workflows")
	}
	return wfs, nil
}

func getWorkflowActions(context context.Context, db db.Database, wfID string) (*pb.WorkflowActionList, error) {
	actions, err := db.GetWorkflowActions(context, wfID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "workflow_id is invalid")
	}
	return actions, nil
}

// The below function check whether a particular workflow context is applicable or needed to
// be send to a worker based on the state of the current action and the targeted workerID.
func isApplicableToSend(context context.Context, wfContext *pb.WorkflowContext, workerID string, db db.Database) bool {
	if wfContext.GetCurrentActionState() == pb.ActionState_ACTION_FAILED ||
		wfContext.GetCurrentActionState() == pb.ActionState_ACTION_TIMEOUT {
		return false
	}
	actions, err := getWorkflowActions(context, db, wfContext.GetWorkflowId())
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
