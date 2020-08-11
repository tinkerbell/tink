package grpcserver

import (
	"context"
	"time"

	"github.com/tinkerbell/tink/db"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var workflowData = make(map[string]int)

const (
	errInvalidWorkerID       = "invalid worker id"
	errInvalidWorkflowID     = "invalid workflow id"
	errInvalidTaskName       = "invalid task name"
	errInvalidActionName     = "invalid action name"
	errInvalidTaskReported   = "reported task name does not match the current action details"
	errInvalidActionReported = "reported action name does not match the current action details"

	msgReceivedStatus   = "received action status: %s"
	msgCurrentWfContext = "current workflow context: %s"
	msgSendWfContext    = "send workflow context: %s"
)

// GetWorkflowContexts implements tinkerbell.GetWorkflowContexts
func (s *server) GetWorkflowContexts(req *pb.WorkflowContextRequest, stream pb.WorkflowSvc_GetWorkflowContextsServer) error {
	wfs, err := getWorkflowsForWorker(s.db, req.WorkerId)
	if err != nil {
		return err
	}
	for _, wf := range wfs {
		wfContext, err := s.db.GetWorkflowContexts(context.Background(), wf)
		if err != nil {
			return status.Errorf(codes.Aborted, err.Error())
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
			return nil, status.Errorf(codes.Aborted, err.Error())
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
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	return getWorkflowActions(context, s.db, wfID)
}

// ReportActionStatus implements tinkerbell.ReportActionStatus
func (s *server) ReportActionStatus(context context.Context, req *pb.WorkflowActionStatus) (*pb.Empty, error) {
	wfID := req.GetWorkflowId()
	if len(wfID) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	if len(req.GetTaskName()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidTaskName)
	}
	if len(req.GetActionName()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidActionName)
	}

	logger.Info(msgReceivedStatus, req)

	wfContext, err := s.db.GetWorkflowContexts(context, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	wfActions, err := s.db.GetWorkflowActions(context, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	actionIndex := wfContext.GetCurrentActionIndex()
	if req.GetActionStatus() == pb.ActionState_ACTION_IN_PROGRESS {
		if wfContext.GetCurrentAction() != "" {
			actionIndex = actionIndex + 1
		}
	}
	action := wfActions.ActionList[actionIndex]
	if action.GetTaskName() != req.GetTaskName() {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidTaskReported)
	}
	if action.GetName() != req.GetActionName() {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidActionReported)
	}
	wfContext.CurrentWorker = action.GetWorkerId()
	wfContext.CurrentTask = req.GetTaskName()
	wfContext.CurrentAction = req.GetActionName()
	wfContext.CurrentActionState = req.GetActionStatus()
	wfContext.CurrentActionIndex = actionIndex
	err = s.db.UpdateWorkflowState(context, wfContext)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Aborted, err.Error())
	}

	// TODO the below "time" would be a part of the request which is coming form worker.
	time := time.Now()
	err = s.db.InsertIntoWorkflowEventTable(context, req, time)
	if err != nil {
		return &pb.Empty{}, status.Error(codes.Aborted, err.Error())
	}
	logger.Info(msgCurrentWfContext, wfContext)
	return &pb.Empty{}, nil
}

// Update Workflow Ephemeral Data
func (s *server) UpdateWorkflowData(context context.Context, req *pb.UpdateWorkflowDataRequest) (*pb.Empty, error) {
	wfID := req.GetWorkflowID()
	if len(wfID) == 0 {
		return &pb.Empty{}, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	_, ok := workflowData[wfID]
	if !ok {
		workflowData[wfID] = 1
	}
	err := s.db.InsertIntoWfDataTable(context, req)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.Empty{}, nil
}

// Get Workflow Ephemeral Data
func (s *server) GetWorkflowData(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	wfID := req.GetWorkflowID()
	if len(wfID) == 0 {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	data, err := s.db.GetfromWfDataTable(context, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow
func (s *server) GetWorkflowMetadata(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	data, err := s.db.GetWorkflowMetadata(context, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowDataVersion returns the latest version of data for a workflow
func (s *server) GetWorkflowDataVersion(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	version, err := s.db.GetWorkflowDataVersion(context, req.WorkflowID)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Version: version}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Version: version}, nil
}

func getWorkflowsForWorker(db db.Database, id string) ([]string, error) {
	if len(id) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkerID)
	}
	wfs, err := db.GetWorkflowsForWorker(id)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	return wfs, nil
}

func getWorkflowActions(context context.Context, db db.Database, wfID string) (*pb.WorkflowActionList, error) {
	actions, err := db.GetWorkflowActions(context, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, errInvalidWorkflowID)
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
				logger.Info(msgSendWfContext, wfContext.GetWorkflowId())
				return true
			}
		}
	} else {
		if actions.ActionList[wfContext.GetCurrentActionIndex()].GetWorkerId() == workerID {
			logger.Info(msgSendWfContext, wfContext.GetWorkflowId())
			return true
		}
	}
	return false
}

func isLastAction(wfContext *pb.WorkflowContext, actions *pb.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}
