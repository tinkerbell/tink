package server

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetWorkflowContexts implements tinkerbell.GetWorkflowContexts.
func (s *DBServer) GetWorkflowContexts(req *workflow.WorkflowContextRequest, stream workflow.WorkflowService_GetWorkflowContextsServer) error {
	wfs, err := getWorkflowsForWorker(stream.Context(), s.db, req.WorkerId)
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

// GetWorkflowContextList implements tinkerbell.GetWorkflowContextList.
func (s *DBServer) GetWorkflowContextList(ctx context.Context, req *workflow.WorkflowContextRequest) (*workflow.WorkflowContextList, error) {
	wfs, err := getWorkflowsForWorker(ctx, s.db, req.WorkerId)
	if err != nil {
		return nil, err
	}

	if wfs != nil {
		wfContexts := []*workflow.WorkflowContext{}
		for _, wf := range wfs {
			wfContext, err := s.db.GetWorkflowContexts(ctx, wf)
			if err != nil {
				return nil, status.Errorf(codes.Aborted, err.Error())
			}
			wfContexts = append(wfContexts, wfContext)
		}
		return &workflow.WorkflowContextList{
			WorkflowContexts: wfContexts,
		}, nil
	}
	return nil, nil
}

// GetWorkflowActions implements tinkerbell.GetWorkflowActions.
func (s *DBServer) GetWorkflowActions(ctx context.Context, req *workflow.WorkflowActionsRequest) (*workflow.WorkflowActionList, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	return getWorkflowActions(ctx, s.db, wfID)
}

// ReportActionStatus implements tinkerbell.ReportActionStatus.
func (s *DBServer) ReportActionStatus(ctx context.Context, req *workflow.WorkflowActionStatus) (*workflow.Empty, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	if req.GetTaskName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidTaskName)
	}
	if req.GetActionName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidActionName)
	}

	l := s.logger.With("actionName", req.GetActionName(), "workflowID", req.GetWorkflowId(), "taskName", req.GetTaskName())
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
	if req.GetActionStatus() == workflow.State_STATE_RUNNING {
		if wfContext.GetCurrentAction() != "" {
			actionIndex++
		}
	}

	if wfContext.TotalNumberOfActions > 1 && actionIndex == wfContext.TotalNumberOfActions-1 {
		return nil, status.Errorf(codes.FailedPrecondition, errInvalidActionIndex)
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
		return &workflow.Empty{}, status.Errorf(codes.Aborted, err.Error())
	}

	// TODO the below "time" would be a part of the request which is coming form worker.
	t := time.Now()
	err = s.db.InsertIntoWorkflowEventTable(ctx, req, t)
	if err != nil {
		return &workflow.Empty{}, status.Error(codes.Aborted, err.Error())
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
	return &workflow.Empty{}, nil
}

// UpdateWorkflowData updates workflow ephemeral data.
func (s *DBServer) UpdateWorkflowData(ctx context.Context, req *workflow.UpdateWorkflowDataRequest) (*workflow.Empty, error) {
	if wfID := req.GetWorkflowId(); wfID == "" {
		return &workflow.Empty{}, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}

	err := s.db.InsertIntoWfDataTable(ctx, req)
	if err != nil {
		return &workflow.Empty{}, status.Errorf(codes.Aborted, err.Error())
	}
	return &workflow.Empty{}, nil
}

// GetWorkflowData gets the ephemeral data for a workflow.
func (s *DBServer) GetWorkflowData(ctx context.Context, req *workflow.GetWorkflowDataRequest) (*workflow.GetWorkflowDataResponse, error) {
	if id := req.GetWorkflowId(); id == "" {
		return &workflow.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}

	data, err := s.db.GetfromWfDataTable(ctx, req)
	if err != nil {
		s.logger.Error(err, "Error getting from data table")
		return &workflow.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Aborted, err.Error())
	}
	return &workflow.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow.
func (s *DBServer) GetWorkflowMetadata(ctx context.Context, req *workflow.GetWorkflowDataRequest) (*workflow.GetWorkflowDataResponse, error) {
	data, err := s.db.GetWorkflowMetadata(ctx, req)
	if err != nil {
		return &workflow.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Aborted, err.Error())
	}
	return &workflow.GetWorkflowDataResponse{Data: data}, nil
}

// GetWorkflowDataVersion returns the latest version of data for a workflow.
func (s *DBServer) GetWorkflowDataVersion(ctx context.Context, req *workflow.GetWorkflowDataRequest) (*workflow.GetWorkflowDataResponse, error) {
	version, err := s.db.GetWorkflowDataVersion(ctx, req.WorkflowId)
	if err != nil {
		return &workflow.GetWorkflowDataResponse{Version: version}, status.Errorf(codes.Aborted, err.Error())
	}
	return &workflow.GetWorkflowDataResponse{Version: version}, nil
}

func getWorkflowsForWorker(ctx context.Context, d db.Database, id string) ([]string, error) {
	if id == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkerID)
	}
	wfs, err := d.GetWorkflowsForWorker(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	return wfs, nil
}

func getWorkflowActions(ctx context.Context, d db.Database, wfID string) (*workflow.WorkflowActionList, error) {
	actions, err := d.GetWorkflowActions(ctx, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, errInvalidWorkflowID)
	}
	return actions, nil
}

// isApplicableToSend checks if a particular workflow context is applicable or if it is needed to
// be sent to a worker based on the state of the current action and the targeted workerID.
func isApplicableToSend(ctx context.Context, logger log.Logger, wfContext *workflow.WorkflowContext, workerID string, d db.Database) bool {
	if wfContext.GetCurrentActionState() == workflow.State_STATE_FAILED ||
		wfContext.GetCurrentActionState() == workflow.State_STATE_TIMEOUT {
		return false
	}
	actions, err := getWorkflowActions(ctx, d, wfContext.GetWorkflowId())
	if err != nil {
		return false
	}
	if wfContext.GetCurrentActionState() == workflow.State_STATE_SUCCESS {
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

func isLastAction(wfContext *workflow.WorkflowContext, actions *workflow.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}
