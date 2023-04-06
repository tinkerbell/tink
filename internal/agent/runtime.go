package agent

import (
	"context"

	"github.com/tinkerbell/tink/internal/agent/workflow"
)

// ContainerRuntime is a runtime capable of executing workflow actions.
type ContainerRuntime interface {
	// Run executes the action. The runtime should mount the following files for the action
	// implementation to communicate a reason and message in the event of failure:
	//
	//	/tinkerbell/failure-reason
	//	/tinkerbell/failure-message
	//
	// The reason and message should be communicataed via the returned error. The message should
	// be the error message and the reason should be provided as defined in failure.Reason().
	Run(context.Context, workflow.Action) error
}
