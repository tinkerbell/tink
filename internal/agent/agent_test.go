package agent_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"github.com/tinkerbell/tink/internal/agent"
	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/failure"
	"github.com/tinkerbell/tink/internal/agent/runtime"
	"github.com/tinkerbell/tink/internal/agent/transport"
	"github.com/tinkerbell/tink/internal/agent/workflow"
	"go.uber.org/zap"
)

func TestAgent_InvalidStart(t *testing.T) {
	cases := []struct {
		Name  string
		Agent *agent.Agent
		Error string
	}{
		{
			Name: "MissingAgentID",
			Agent: &agent.Agent{
				Log:       logr.Discard(),
				Transport: transport.Noop(),
				Runtime:   runtime.Noop(),
			},
			Error: "ID field must be set before calling Start()",
		},
		{
			Name: "MissingRuntime",
			Agent: &agent.Agent{
				Log:       logr.Discard(),
				ID:        "1234",
				Transport: transport.Noop(),
			},
			Error: "Runtime field must be set before calling Start()",
		},
		{
			Name: "MissingTransport",
			Agent: &agent.Agent{
				Log:     logr.Discard(),
				ID:      "1234",
				Runtime: runtime.Noop(),
			},
			Error: "Transport field must be set before calling Start()",
		},
		{
			Name: "InitializedCorrectly",
			Agent: &agent.Agent{
				Log:       logr.Discard(),
				ID:        "1234",
				Transport: transport.Noop(),
				Runtime:   runtime.Noop(),
			},
		},
		{
			Name: "NoLogger",
			Agent: &agent.Agent{
				ID:        "1234",
				Transport: transport.Noop(),
				Runtime:   runtime.Noop(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Agent.Start(context.Background())
			switch {
			case tc.Error != "" && err == nil:
				t.Fatalf("Expected error '%v' but received none", tc.Error)
			case tc.Error != "" && err != nil && !strings.Contains(err.Error(), tc.Error):
				t.Fatalf("Expected: %v\n;\nReceived: %v", tc.Error, err)
			case tc.Error == "" && err != nil:
				t.Fatalf("Received unexpected error: %v", err)
			}
		})
	}
}

// The goal of this test is to ensure the agent rejects concurrent workflows.
func TestAgent_ConcurrentWorkflows(t *testing.T) {
	logger := zapr.NewLogger(zap.Must(zap.NewDevelopment()))
	recorder := event.NoopRecorder()
	trnport := transport.Noop()

	wflw := workflow.Workflow{
		ID: "1234",
		Actions: []workflow.Action{
			{
				ID:    "1234",
				Name:  "name",
				Image: "image",
			},
		},
	}

	// Started is used to indicate the runtime has received the workflow.
	started := make(chan struct{})
	rntime := agent.ContainerRuntimeMock{
		RunFunc: func(ctx context.Context, action workflow.Action) error {
			started <- struct{}{}
			<-ctx.Done()
			return nil
		},
	}

	agnt := agent.Agent{
		Log:       logger,
		Transport: &trnport,
		Runtime:   &rntime,
		ID:        "1234",
	}

	// HandleWorkflow will reject us if we haven't started the agent first.
	if err := agnt.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Build a cancellable context so we can tear everything down. The timeout is guesswork but
	// this test shouldn't take long.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Handle the first workflow and wait for it to start.
	agnt.HandleWorkflow(ctx, wflw, recorder)

	// Wait for the container runtime to start.
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	// Attempt to fire off a second workflow.
	agnt.HandleWorkflow(ctx, wflw, recorder)

	// Ensure the latest event recorded is a event.WorkflowRejected.
	calls := recorder.RecordEventCalls()
	if len(calls) == 0 {
		t.Fatal("No calls have been made to the event.Recorder")
	}

	lastCall := calls[len(calls)-1]
	ev, ok := lastCall.Event.(event.WorkflowRejected)
	if !ok {
		t.Fatalf("Expected event of type event.WorkflowRejected; received %T", ev)
	}

	expectEvent := event.WorkflowRejected{
		ID:      wflw.ID,
		Message: "workflow already in progress",
	}
	if !cmp.Equal(expectEvent, ev) {
		t.Fatalf("Received unexpected event:\n%v", cmp.Diff(expectEvent, ev))
	}
}

func TestAgent_HandlingWorkflows(t *testing.T) {
	logger := zapr.NewLogger(zap.Must(zap.NewDevelopment()))

	type ReasonAndMessage struct {
		Reason, Message string
	}

	cases := []struct {
		Name     string
		Workflow workflow.Workflow
		Errors   map[string]ReasonAndMessage
		Events   []event.Event
	}{
		{
			Name: "SuccessfulWorkflow",
			Workflow: workflow.Workflow{
				ID: "1234",
				Actions: []workflow.Action{
					{
						ID:    "1",
						Name:  "action_1",
						Image: "image_1",
					},
					{
						ID:    "2",
						Name:  "action_2",
						Image: "image_2",
					},
				},
			},
			Errors: map[string]ReasonAndMessage{},
			Events: []event.Event{
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionSucceeded{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "2",
				},
				event.ActionSucceeded{
					WorkflowID: "1234",
					ActionID:   "2",
				},
			},
		},
		{
			Name: "LastActionFails",
			Workflow: workflow.Workflow{
				ID: "1234",
				Actions: []workflow.Action{
					{
						ID:    "1",
						Name:  "action_1",
						Image: "image_1",
					},
					{
						ID:    "2",
						Name:  "action_2",
						Image: "image_2",
					},
				},
			},
			Errors: map[string]ReasonAndMessage{
				"2": {
					Reason:  "TestReason",
					Message: "test message",
				},
			},
			Events: []event.Event{
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionSucceeded{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "2",
				},
				event.ActionFailed{
					WorkflowID: "1234",
					ActionID:   "2",
					Reason:     "TestReason",
					Message:    "test message",
				},
			},
		},
		{
			Name: "FirstActionFails",
			Workflow: workflow.Workflow{
				ID: "1234",
				Actions: []workflow.Action{
					{
						ID:    "1",
						Name:  "action_1",
						Image: "image_1",
					},
					{
						ID:    "2",
						Name:  "action_2",
						Image: "image_2",
					},
				},
			},
			Errors: map[string]ReasonAndMessage{
				"1": {
					Reason:  "TestReason",
					Message: "test message",
				},
			},
			Events: []event.Event{
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionFailed{
					WorkflowID: "1234",
					ActionID:   "1",
					Reason:     "TestReason",
					Message:    "test message",
				},
			},
		},
		{
			Name: "MiddleActionFails",
			Workflow: workflow.Workflow{
				ID: "1234",
				Actions: []workflow.Action{
					{
						ID:    "1",
						Name:  "action_1",
						Image: "image_1",
					},
					{
						ID:    "2",
						Name:  "action_2",
						Image: "image_2",
					},
					{
						ID:    "3",
						Name:  "action_3",
						Image: "image_3",
					},
				},
			},
			Errors: map[string]ReasonAndMessage{
				"2": {
					Reason:  "TestReason",
					Message: "test message",
				},
			},
			Events: []event.Event{
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionSucceeded{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "2",
				},
				event.ActionFailed{
					WorkflowID: "1234",
					ActionID:   "2",
					Reason:     "TestReason",
					Message:    "test message",
				},
			},
		},
		{
			Name: "InvalidReason",
			Workflow: workflow.Workflow{
				ID: "1234",
				Actions: []workflow.Action{
					{
						ID:    "1",
						Name:  "action_1",
						Image: "image_1",
					},
				},
			},
			Errors: map[string]ReasonAndMessage{
				"1": {
					Reason:  "this is an invalid reason format",
					Message: "test message",
				},
			},
			Events: []event.Event{
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionFailed{
					WorkflowID: "1234",
					ActionID:   "1",
					Reason:     "InvalidReason",
					Message:    "test message",
				},
			},
		},
		{
			Name: "InvalidMessage",
			Workflow: workflow.Workflow{
				ID: "1234",
				Actions: []workflow.Action{
					{
						ID:    "1",
						Name:  "action_1",
						Image: "image_1",
					},
				},
			},
			Errors: map[string]ReasonAndMessage{
				"1": {
					Reason: "TestReason",
					Message: `invalid 
message`,
				},
			},
			Events: []event.Event{
				event.ActionStarted{
					WorkflowID: "1234",
					ActionID:   "1",
				},
				event.ActionFailed{
					WorkflowID: "1234",
					ActionID:   "1",
					Reason:     "TestReason",
					Message:    `invalid \nmessage`,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			trnport := transport.Noop()

			rntime := agent.ContainerRuntimeMock{
				RunFunc: func(ctx context.Context, action workflow.Action) error {
					if res, ok := tc.Errors[action.ID]; ok {
						return failure.NewReason(res.Message, res.Reason)
					}
					return nil
				},
			}

			// The event recorder is what tells us the workflow has finished executing so we use it
			// to check for the last expected action.
			lastEventReceived := make(chan struct{})
			recorder := event.RecorderMock{
				RecordEventFunc: func(contextMoqParam context.Context, event event.Event) error {
					if cmp.Equal(event, tc.Events[len(tc.Events)-1]) {
						lastEventReceived <- struct{}{}
					}
					return nil
				},
			}

			// Create and start the agent as start is a prereq to calling HandleWorkflow().
			agnt := agent.Agent{
				Log:       logger,
				Transport: &trnport,
				Runtime:   &rntime,
				ID:        "1234",
			}
			if err := agnt.Start(context.Background()); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Configure a timeout of 5 seconds, this test shouldn't take long.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Handle the workflow
			agnt.HandleWorkflow(ctx, tc.Workflow, &recorder)

			// Wait for the last expected event or timeout.
			select {
			case <-lastEventReceived:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}

			// Validate all events received are what we expected.
			calls := recorder.RecordEventCalls()
			if len(calls) != len(tc.Events) {
				t.Fatalf("Expected %v events; Received %v\n%+v", len(tc.Events), len(calls), calls)
			}

			var receivedEvents []event.Event
			for _, call := range calls {
				receivedEvents = append(receivedEvents, call.Event)
			}

			if !cmp.Equal(tc.Events, receivedEvents) {
				t.Fatalf("Did not received expected event set:\n%v", cmp.Diff(tc.Events, receivedEvents))
			}
		})
	}
}
