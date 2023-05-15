package agent

import (
	"context"
	"errors"
	"sync"

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

	// sem ensure we handle a single workflow at a time.
	sem chan struct{}

	// executionContext tracks the currently executing workflow.
	executionContext *executionContext
	mtx              sync.RWMutex
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
		//nolint:stylecheck // Runtime is a field of agent.
		return errors.New("Runtime field must be set before calling Start()")
	}

	if agent.Log.GetSink() == nil {
		agent.Log = logr.Discard()
	}

	agent.Log = agent.Log.WithValues("agent_id", agent.ID)

	// Initialize the semaphore and add a resource to it ensuring we can run 1 workflow at a time.
	agent.sem = make(chan struct{}, 1)
	agent.sem <- struct{}{}

	return agent.Transport.Start(ctx, agent.ID, agent)
}

// HandleWorkflow satisfies transport.
func (agent *Agent) HandleWorkflow(ctx context.Context, wflw workflow.Workflow, events event.Recorder) {
	if agent.sem == nil {
		agent.Log.Info("Agent must have Start() called before calling HandleWorkflow()")
	}

	select {
	case <-agent.sem:
		// Ensure we configure the current workflow and cancellation func before we launch the
		// goroutine to avoid a race with CancelWorkflow.
		agent.mtx.Lock()
		defer agent.mtx.Unlock()

		ctx, cancel := context.WithCancel(ctx)
		agent.executionContext = &executionContext{
			Workflow: wflw,
			Cancel:   cancel,
		}

		go func() {
			// Replenish the semaphore on exit so we can pick up another workflow.
			defer func() { agent.sem <- struct{}{} }()

			agent.run(ctx, wflw, events)

			// Nilify the execution context after running so cancellation requests are ignored.
			agent.mtx.Lock()
			defer agent.mtx.Unlock()
			agent.executionContext = nil
		}()

	default:
		log := agent.Log.WithValues("workflow_id", wflw.ID)

		reject := event.WorkflowRejected{
			ID:      wflw.ID,
			Message: "workflow already in progress",
		}

		if err := events.RecordEvent(ctx, reject); err != nil {
			log.Error(err, "Failed to record workflow rejection event")
			return
		}

		log.Info("Workflow already executing; dropping request")
	}
}

func (agent *Agent) CancelWorkflow(workflowID string) {
	agent.mtx.RLock()
	defer agent.mtx.RUnlock()

	if agent.executionContext == nil {
		agent.Log.Info("No workflow running; ignoring cancellation request", "workflow_id", workflowID)
		return
	}

	if agent.executionContext.Workflow.ID != workflowID {
		agent.Log.Info(
			"Incorrect workflow ID in cancellation request; ignoring cancellation request",
			"workflow_id", workflowID,
			"running_workflow_id", agent.executionContext.Workflow.ID,
		)
		return
	}

	agent.Log.Info("Cancel workflow", "workflow_id", workflowID)
	agent.executionContext.Cancel()
}

type executionContext struct {
	Workflow workflow.Workflow
	Cancel   context.CancelFunc
}
