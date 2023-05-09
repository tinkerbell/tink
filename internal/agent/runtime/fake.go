package runtime

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/agent"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

var _ agent.ContainerRuntime = Fake{}

func Noop() Fake {
	return Fake{
		Log: logr.Discard(),
	}
}

// Fake is a runtime that always succeeds. It does not literally execute any actions.
type Fake struct {
	Log logr.Logger
}

// Run satisfies agent.ContainerRuntime.
func (f Fake) Run(_ context.Context, a workflow.Action) error {
	f.Log.Info("Starting fake container", "action", a)
	return nil
}
