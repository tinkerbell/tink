package transport

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

type Fake struct {
	Log       logr.Logger
	Workflows []workflow.Workflow
}

func (f Fake) Start(ctx context.Context, _ string, runner WorkflowHandler) error {
	f.Log.Info("Starting fake transport")
	for _, w := range f.Workflows {
		if err := runner.HandleWorkflow(ctx, w, f); err != nil {
			f.Log.Error(err, "Running workflow", "workflow", w)
		}
	}
	return nil
}

func (f Fake) RecordEvent(_ context.Context, e event.Event) error {
	f.Log.Info("Recording event", "event", e.GetName())
	return nil
}
