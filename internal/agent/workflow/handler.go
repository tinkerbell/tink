package workflow

import (
	"context"

	"github.com/tinkerbell/tink/internal/agent/event"
)

// Handler is responsible for handling workflow execution.
type Handler interface {
	// HandleWorkflow begins executing the given workflow. The event recorder can be used to
	// indicate the progress of a workflow. If the given context becomes cancelled, the workflow
	// handler should stop workflow execution.
	HandleWorkflow(context.Context, Workflow, event.Recorder) error
}
