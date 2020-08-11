package mock

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/tinkerbell/tink/db"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

// CreateWorkflow creates a new workflow
func (d DB) CreateWorkflow(ctx context.Context, wf db.Workflow, data string, id uuid.UUID) error {
	return nil
}

// InsertIntoWfDataTable : Insert ephemeral data in workflow_data table
func (d DB) InsertIntoWfDataTable(ctx context.Context, req *pb.UpdateWorkflowDataRequest) error {
	return nil
}

// GetfromWfDataTable : Give you the ephemeral data from workflow_data table
func (d DB) GetfromWfDataTable(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
	if req.WorkflowID == "5711afcf-ea0b-4055-b4d6-9f88080f7afc" {
		return []byte("{'os': 'ubuntu', 'base_url': 'http://192.168.1.1/'}"), nil
	}
	return []byte{}, nil
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow
func (d DB) GetWorkflowMetadata(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
	return []byte{}, nil
}

// GetWorkflowDataVersion returns the latest version of data for a workflow
func (d DB) GetWorkflowDataVersion(ctx context.Context, workflowID string) (int32, error) {
	return int32(0), nil
}

// GetWorkflowsForWorker : returns the list of workflows for a particular worker
func (d DB) GetWorkflowsForWorker(id string) ([]string, error) {
	return []string{}, nil
}

// GetWorkflow returns a workflow
func (d DB) GetWorkflow(ctx context.Context, id string) (db.Workflow, error) {
	return db.Workflow{}, nil
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
	return nil
}

// GetWorkflowContexts : gives you the current workflow context
func (d DB) GetWorkflowContexts(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
	return nil, nil
}

// GetWorkflowActions : gives you the action list of workflow
func (d DB) GetWorkflowActions(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
	return nil, nil
}

// InsertIntoWorkflowEventTable : insert workflow event table
func (d DB) InsertIntoWorkflowEventTable(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
	return nil
}

// ShowWorkflowEvents returns all workflows
func (d DB) ShowWorkflowEvents(wfID string, fn func(wfs *pb.WorkflowActionStatus) error) error {
	return nil
}
