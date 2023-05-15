package transport

import (
	"context"

	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

// WorkflowHandler is responsible for workflow execution.
type WorkflowHandler interface {
	// HandleWorkflow executes the given workflow. The event.Recorder can be used to publish events
	// as the workflow transits its lifecycle. HandleWorkflow should not block and should be efficient
	// in handing off workflow processing.
	HandleWorkflow(context.Context, workflow.Workflow, event.Recorder)

	// CancelWorkflow cancels a workflow identified by workflowID. It should not block and should
	// be efficient in handing off the cancellation request.
	CancelWorkflow(workflowID string)
}
