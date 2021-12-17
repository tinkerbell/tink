package tink

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

type Wdata struct {
	Data WorkflowData
}

type WorkflowData struct {
	WorkerID     string
	WorkflowID   string
	ActionStatus *pb.WorkflowActionStatus
}

// GetWorkflowContexts gets workflow context for worker id on success else returns error.
func (w Wdata) GetWorkflowContexts(ctx context.Context, client pb.WorkflowServiceClient) (pb.WorkflowService_GetWorkflowContextsClient, error) {
	if w.Data.WorkerID == "" {
		return nil, errors.New("Empty string is not a valid worker id")
	}

	if client == nil || (reflect.ValueOf(client).IsNil()) {
		return nil, errors.New("WorkflowServiceClient interface is not valid")
	}

	response, err := client.GetWorkflowContexts(ctx, &pb.WorkflowContextRequest{WorkerId: w.Data.WorkerID})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get workflow contexts")
	}
	return response, nil
}

// GetWorkflowActions gets workflow action list for workflow id on success else returns error.
func (w Wdata) GetWorkflowActions(ctx context.Context, client pb.WorkflowServiceClient) (*pb.WorkflowActionList, error) {
	if w.Data.WorkflowID == "" {
		return nil, errors.New("Empty string is not a valid workflow id")
	}

	if client == nil || (reflect.ValueOf(client).IsNil()) {
		return nil, errors.New("WorkflowServiceClient interface is not valid")
	}

	actions, err := client.GetWorkflowActions(ctx, &pb.WorkflowActionsRequest{WorkflowId: w.Data.WorkflowID})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get workflow actions")
	}
	return actions, nil
}

// ReportActionStatus reports action status on success else returns error.
func (w Wdata) ReportActionStatus(ctx context.Context, client pb.WorkflowServiceClient) error {
	if client == nil || (reflect.ValueOf(client).IsNil()) {
		return errors.New("WorkflowServiceClient interface is not valid")
	}

	if w.Data.ActionStatus == nil || (reflect.ValueOf(w.Data.ActionStatus).IsNil()) {
		return errors.New("WorkflowActionStatus is not valid")
	}

	_, err := client.ReportActionStatus(ctx, w.Data.ActionStatus)
	if err != nil {
		return errors.Wrap(err, "Failed to report action status")
	}
	return err
}
