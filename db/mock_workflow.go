package db

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

// MockDB is the mocked implementation of Database interface
type MockDB struct{}

// CreateWorkflow creates a new workflow
func (mdb MockDB) CreateWorkflow(ctx context.Context, wf Workflow, data string, id uuid.UUID) error {
	return nil
}

// InsertIntoWfDataTable : Insert ephemeral data in workflow_data table
func (mdb MockDB) InsertIntoWfDataTable(ctx context.Context, req *pb.UpdateWorkflowDataRequest) error {
	return nil
}

// GetfromWfDataTable : Give you the ephemeral data from workflow_data table
func (mdb MockDB) GetfromWfDataTable(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
	if req.WorkflowID == "5711afcf-ea0b-4055-b4d6-9f88080f7afc" {
		return []byte("some workflow data"), nil
	}
	return []byte{}, nil
}

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow
func (mdb MockDB) GetWorkflowMetadata(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
	return []byte{}, nil
}

// GetWorkflowDataVersion returns the latest version of data for a workflow
func (mdb MockDB) GetWorkflowDataVersion(ctx context.Context, workflowID string) (int32, error) {
	return int32(0), nil
}

// GetWorkflowsForWorker : returns the list of workflows for a particular worker
func (mdb MockDB) GetWorkflowsForWorker(id string) ([]string, error) {
	return []string{}, nil
}

// GetWorkflow returns a workflow
func (mdb MockDB) GetWorkflow(ctx context.Context, id string) (Workflow, error) {
	return Workflow{}, nil
}

// DeleteWorkflow deletes a workflow
func (mdb MockDB) DeleteWorkflow(ctx context.Context, id string, state int32) error {
	return nil
}

// ListWorkflows returns all workflows
func (mdb MockDB) ListWorkflows(fn func(wf Workflow) error) error {
	return nil
}

// UpdateWorkflow updates a given workflow
func (mdb MockDB) UpdateWorkflow(ctx context.Context, wf Workflow, state int32) error {
	return nil
}

// UpdateWorkflowState : update the current workflow state
func (mdb MockDB) UpdateWorkflowState(ctx context.Context, wfContext *pb.WorkflowContext) error {
	return nil
}

// GetWorkflowContexts : gives you the current workflow context
func (mdb MockDB) GetWorkflowContexts(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
	return nil, nil
}

// GetWorkflowActions : gives you the action list of workflow
func (mdb MockDB) GetWorkflowActions(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
	return nil, nil
}

// InsertIntoWorkflowEventTable : insert workflow event table
func (mdb MockDB) InsertIntoWorkflowEventTable(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
	return nil
}

// ShowWorkflowEvents returns all workflows
func (mdb MockDB) ShowWorkflowEvents(wfID string, fn func(wfs *pb.WorkflowActionStatus) error) error {
	return nil
}
