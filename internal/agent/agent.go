package agent

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

// Agent is the core data structure for handling workflow execution on target nodes. It leverages
// a Transport and a ContainerRuntime to retrieve workflows and execute actions.
//
// The agent runs a single workflow at a time. Concurrent requests to run workflows will have the
// second workflow rejected with an event.WorkflowRejected event.
type Agent struct {
	Log logr.Logger

	// ID is the unique identifier for the agent. It is used by the transport to identify workflows
	// scheduled for this agent.
	ID string

	// Transport is the transport used by the agent for communicating workflows and events.
	Transport Transport

	// Runtime is the container runtime used to execute workflow actions.
	Runtime ContainerRuntime

	sem chan struct{}
}

// Start finalizes the Agent configuration and starts the configured Transport so it is ready
// to receive workflows. On receiving a workflow, it will leverage the configured Runtime to
// execute workflow actions.
func (agent *Agent) Start(ctx context.Context) error {
	if agent.ID == "" {
		return errors.New("ID field must be set before calling Start()")
	}

	if agent.Transport == nil {
		return errors.New("Transport field must be set before calling Start()")
	}

	if agent.Runtime == nil {
		//nolint:stylecheck // Specifying field on data structure
		return errors.New("Runtime field must be set before calling Start()")
	}

	agent.Log = agent.Log.WithValues("agent_id", agent.ID)

	// Initialize the semaphore and add a resource to it ensuring we can run 1 workflow at a time.
	agent.sem = make(chan struct{}, 1)
	agent.sem <- struct{}{}

	agent.Log.Info("Starting agent")
	return agent.Transport.Start(ctx, agent.ID, agent)
}

// HandleWorkflow satisfies transport.
func (agent *Agent) HandleWorkflow(ctx context.Context, wflw workflow.Workflow, events event.Recorder) error {
	// sem isn't protected by a synchronization data structure so this is technically invoking
	// undefined behavior - consider this a best effort to ensuring Start() has been called.
	if agent.sem == nil {
		return errors.New("agent must have Start() called before calling HandleWorkflow()")
	}

	select {
	case <-agent.sem:
		// Replenish the semaphore on exit so we can pick up another workflow.
		defer func() { agent.sem <- struct{}{} }()
		return agent.run(ctx, wflw, events)

	default:
		reject := event.WorkflowRejected{
			ID:      wflw.ID,
			Message: "workflow already in progress",
		}
		events.RecordEvent(ctx, reject)
		return nil
	}
}
