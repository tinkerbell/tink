package mock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tinkerbell/tink/db"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

// DB is the mocked implementation of Database interface
type DB struct {
	// workflow
	CreateWorkflowFunc               func(ctx context.Context, wf db.Workflow, data string, id uuid.UUID) error
	GetfromWfDataTableFunc           func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error)
	InsertIntoWfDataTableFunc        func(ctx context.Context, req *pb.UpdateWorkflowDataRequest) error
	GetWorkflowMetadataFunc          func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error)
	GetWorkflowDataVersionFunc       func(ctx context.Context, workflowID string) (int32, error)
	GetWorkflowsForWorkerFunc        func(id string) ([]string, error)
	GetWorkflowContextsFunc          func(ctx context.Context, wfID string) (*pb.WorkflowContext, error)
	GetWorkflowActionsFunc           func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error)
	UpdateWorkflowStateFunc          func(ctx context.Context, wfContext *pb.WorkflowContext) error
	InsertIntoWorkflowEventTableFunc func(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error
	// template
	GetTemplateFunc    func(ctx context.Context, id string) (string, string, error)
	CreateTemplateFunc func(ctx context.Context, name string, data string, id uuid.UUID) error
}
