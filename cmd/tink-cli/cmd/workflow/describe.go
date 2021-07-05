package workflow

import (
	"context"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/describe"
	"github.com/tinkerbell/tink/protos/workflow"
)

type describeWorkflow struct {
	describe.Options
}

func (d *describeWorkflow) DescribeByID(ctx context.Context, cl *client.FullClient, requestedID string) (interface{}, error) {
	return cl.WorkflowClient.GetWorkflow(ctx, &workflow.GetRequest{Id: requestedID})
}

func NewDescribeOptions() describe.Options {
	w := describeWorkflow{}
	return describe.Options{
		DescribeByID: w.DescribeByID,
	}
}
