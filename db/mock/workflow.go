package mock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tinkerbell/tink/db"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

// CreateWorkflow creates a new workflow
func (d DB) CreateWorkflow(ctx context.Context, wf db.Workflow, data string, id uuid.UUID) error {
	return d.CreateWorkflowFunc(ctx, wf, data, id)
}

// InsertIntoWfDataTable : Insert ephemeral data in workflow_data table
func (d DB) InsertIntoWfDataTable(ctx context.Context, req *pb.UpdateWorkflowDataRequest) error {
	return d.InsertIntoWfDataTableFunc(ctx, req)
}

// GetfromWfDataTable : Give you the ephemeral data from workflow_data table
func (d DB) GetfromWfDataTable(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
	return d.GetfromWfDataTableFunc(ctx, req)
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow
func (d DB) GetWorkflowMetadata(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
	return d.GetWorkflowMetadataFunc(ctx, req)
}

// GetWorkflowDataVersion returns the latest version of data for a workflow
func (d DB) GetWorkflowDataVersion(ctx context.Context, workflowID string) (int32, error) {
	return d.GetWorkflowDataVersionFunc(ctx, workflowID)
}

// GetWorkflowsForWorker : returns the list of workflows for a particular worker
func (d DB) GetWorkflowsForWorker(ctx context.Context, id string) ([]string, error) {
	return d.GetWorkflowsForWorkerFunc(ctx, id)
}

// GetWorkflow returns a workflow
func (d DB) GetWorkflow(ctx context.Context, id string) (db.Workflow, error) {
	return d.GetWorkflowFunc(ctx, id)
}

// DeleteWorkflow deletes a workflow
func (d DB) DeleteWorkflow(ctx context.Context, id string, state int32) error {
	return nil
}

// ListWorkflows returns all workflows
func (d DB) ListWorkflows(fn func(wf db.Workflow) error) error {
	return nil
}

// UpdateWorkflow updates a given workflow
func (d DB) UpdateWorkflow(ctx context.Context, wf db.Workflow, state int32) error {
	return nil
}

// UpdateWorkflowState : update the current workflow state
func (d DB) UpdateWorkflowState(ctx context.Context, wfContext *pb.WorkflowContext) error {
	return d.UpdateWorkflowStateFunc(ctx, wfContext)
}

// GetWorkflowContexts : gives you the current workflow context
func (d DB) GetWorkflowContexts(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
	return d.GetWorkflowContextsFunc(ctx, wfID)
}

// GetWorkflowActions : gives you the action list of workflow
func (d DB) GetWorkflowActions(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
	return d.GetWorkflowActionsFunc(ctx, wfID)
}

// InsertIntoWorkflowEventTable : insert workflow event table
func (d DB) InsertIntoWorkflowEventTable(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
	return d.InsertIntoWorkflowEventTableFunc(ctx, wfEvent, time)
}

// ShowWorkflowEvents returns all workflows
func (d DB) ShowWorkflowEvents(wfID string, fn func(wfs *pb.WorkflowActionStatus) error) error {
	return nil
}
