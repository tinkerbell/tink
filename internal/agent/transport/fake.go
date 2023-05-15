package transport

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

func Noop() Fake {
	return Fake{
		Log: logr.Discard(),
	}
}

type Fake struct {
	Log       logr.Logger
	Workflows []workflow.Workflow
}

func (f Fake) Start(ctx context.Context, _ string, handler WorkflowHandler) error {
	f.Log.Info("Starting fake transport")
	for _, w := range f.Workflows {
		handler.HandleWorkflow(ctx, w, f)
	}
	return nil
}

func (f Fake) RecordEvent(_ context.Context, e event.Event) error {
	f.Log.Info("Recording event", "event", e.GetName())
	return nil
}
