package grpcserver

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/db"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var workflowData = make(map[string]int)

const (
	errInvalidWorkerID       = "invalid worker id"
	errInvalidWorkflowId     = "invalid workflow id"
	errInvalidTaskName       = "invalid task name"
	errInvalidActionName     = "invalid action name"
	errInvalidTaskReported   = "reported task name does not match the current action details"
	errInvalidActionReported = "reported action name does not match the current action details"

	msgReceivedStatus   = "received action status: %s"
	msgCurrentWfContext = "current workflow context"
	msgSendWfContext    = "send workflow context: %s"
)

// GetWorkflowContexts implements tinkerbell.GetWorkflowContexts
func (s *server) GetWorkflowContexts(req *pb.WorkflowContextRequest, stream pb.WorkflowService_GetWorkflowContextsServer) error {
	wfs, err := getWorkflowsForWorker(s.db, req.WorkerId)
	if err != nil {
		return err
	}
	for _, wf := range wfs {
		wfContext, err := s.db.GetWorkflowContexts(stream.Context(), wf)
		if err != nil {
			return status.Errorf(codes.Aborted, err.Error())
		}
		if isApplicableToSend(stream.Context(), s.logger, wfContext, req.WorkerId, s.db) {
			if err := stream.Send(wfContext); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetWorkflowContextList implements tinkerbell.GetWorkflowContextList
	wfs, err := getWorkflowsForWorker(s.db, req.WorkerId)
func (s *server) GetWorkflowContextList(ctx context.Context, req *pb.WorkflowContextRequest) (*pb.WorkflowContextList, error) {
	if err != nil {
		return nil, err
	}

	if wfs != nil {
		wfContexts := []*pb.WorkflowContext{}
		for _, wf := range wfs {
			wfContext, err := s.db.GetWorkflowContexts(ctx, wf)
			if err != nil {
				return nil, status.Errorf(codes.Aborted, err.Error())
			}
			wfContexts = append(wfContexts, wfContext)
		}
		return &pb.WorkflowContextList{
			WorkflowContexts: wfContexts,
		}, nil
	}
	return nil, nil
}

// GetWorkflowActions implements tinkerbell.GetWorkflowActions
func (s *server) GetWorkflowActions(ctx context.Context, req *pb.WorkflowActionsRequest) (*pb.WorkflowActionList, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowId)
	}
	return getWorkflowActions(ctx, s.db, wfID)
}

// ReportActionStatus implements tinkerbell.ReportActionStatus
func (s *server) ReportActionStatus(ctx context.Context, req *pb.WorkflowActionStatus) (*pb.Empty, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowId)
	}
	if req.GetTaskName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidTaskName)
	}
	if req.GetActionName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidActionName)
	}

	l := s.logger.With("actionName", req.GetActionName(), "workflowID", req.GetWorkflowId())
	l.Info(fmt.Sprintf(msgReceivedStatus, req.GetActionStatus()))

	wfContext, err := s.db.GetWorkflowContexts(ctx, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	wfActions, err := s.db.GetWorkflowActions(ctx, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	actionIndex := wfContext.GetCurrentActionIndex()
	if req.GetActionStatus() == pb.State_STATE_RUNNING {
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
	err = s.db.UpdateWorkflowState(ctx, wfContext)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Aborted, err.Error())
	}

	// TODO the below "time" would be a part of the request which is coming form worker.
	time := time.Now()
	err = s.db.InsertIntoWorkflowEventTable(ctx, req, time)
	if err != nil {
		return &pb.Empty{}, status.Error(codes.Aborted, err.Error())
	}

	l = s.logger.With(
		"workflowID", wfContext.GetWorkflowId(),
		"currentWorker", wfContext.GetCurrentWorker(),
		"currentTask", wfContext.GetCurrentTask(),
		"currentAction", wfContext.GetCurrentAction(),
		"currentActionIndex", strconv.FormatInt(wfContext.GetCurrentActionIndex(), 10),
		"currentActionState", wfContext.GetCurrentActionState(),
		"totalNumberOfActions", wfContext.GetTotalNumberOfActions(),
	)
	l.Info(msgCurrentWfContext)
	return &pb.Empty{}, nil
}

// UpdateWorkflowData updates workflow ephemeral data
func (s *server) UpdateWorkflowData(ctx context.Context, req *pb.UpdateWorkflowDataRequest) (*pb.Empty, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return &pb.Empty{}, status.Errorf(codes.InvalidArgument, errInvalidWorkflowId)
	}
	_, ok := workflowData[wfID]
	if !ok {
		workflowData[wfID] = 1
	}
	err := s.db.InsertIntoWfDataTable(ctx, req)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.Empty{}, nil
}

// GetWorkflowData gets the ephemeral data for a workflow
func (s *server) GetWorkflowData(ctx context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.InvalidArgument, errInvalidWorkflowId)
	}
	data, err := s.db.GetfromWfDataTable(ctx, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow
func (s *server) GetWorkflowMetadata(ctx context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	data, err := s.db.GetWorkflowMetadata(ctx, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowDataVersion returns the latest version of data for a workflow
func (s *server) GetWorkflowDataVersion(ctx context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	version, err := s.db.GetWorkflowDataVersion(ctx, req.WorkflowId)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Version: version}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Version: version}, nil
}

func getWorkflowsForWorker(db db.Database, id string) ([]string, error) {
	if id == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkerID)
	}
	wfs, err := db.GetWorkflowsForWorker(id)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	return wfs, nil
}

func getWorkflowActions(ctx context.Context, db db.Database, wfID string) (*pb.WorkflowActionList, error) {
	actions, err := db.GetWorkflowActions(ctx, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, errInvalidWorkflowId)
	}
	return actions, nil
}

// isApplicableToSend checks if a particular workflow context is applicable or if it is needed to
// be sent to a worker based on the state of the current action and the targeted workerID
func isApplicableToSend(ctx context.Context, logger log.Logger, wfContext *pb.WorkflowContext, workerID string, db db.Database) bool {
	if wfContext.GetCurrentActionState() == pb.State_STATE_FAILED ||
		wfContext.GetCurrentActionState() == pb.State_STATE_TIMEOUT {
		return false
	}
	actions, err := getWorkflowActions(ctx, db, wfContext.GetWorkflowId())
	if err != nil {
		return false
	}
	if wfContext.GetCurrentActionState() == pb.State_STATE_SUCCESS {
		if isLastAction(wfContext, actions) {
			return false
		}
		if wfContext.GetCurrentActionIndex() == 0 {
			if actions.ActionList[wfContext.GetCurrentActionIndex()+1].GetWorkerId() == workerID {
				logger.Info(fmt.Sprintf(msgSendWfContext, wfContext.GetWorkflowId()))
				return true
			}
		}
	} else if actions.ActionList[wfContext.GetCurrentActionIndex()].GetWorkerId() == workerID {
		logger.Info(fmt.Sprintf(msgSendWfContext, wfContext.GetWorkflowId()))
		return true

	}
	return false
}

func isLastAction(wfContext *pb.WorkflowContext, actions *pb.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}
