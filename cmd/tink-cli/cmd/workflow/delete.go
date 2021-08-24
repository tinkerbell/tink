package workflow

import (
	"context"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/delete"
	"github.com/tinkerbell/tink/protos/workflow"
)

type deleteWorkflow struct {
	delete.Options
}

func (d *deleteWorkflow) DeleteByID(ctx context.Context, cl *client.FullClient, requestedID string) (interface{}, error) {
	return cl.WorkflowClient.DeleteWorkflow(ctx, &workflow.GetRequest{Id: requestedID})
}

func NewDeleteOptions() delete.Options {
	w := deleteWorkflow{}
	return delete.Options{
		DeleteByID: w.DeleteByID,
	}
}
