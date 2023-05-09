package agent_test

import (
	"context"
	"strings"
	"testing"

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
		Agent agent.Agent
		Error string
	}{
		{
			Name: "MissingAgentID",
			Agent: agent.Agent{
				Log:       logr.Discard(),
				Transport: transport.Noop(),
				Runtime:   runtime.Noop(),
			},
			Error: "ID field must be set before calling Start()",
		},
		{
			Name: "MissingRuntime",
			Agent: agent.Agent{
				Log:       logr.Discard(),
				ID:        "1234",
				Transport: transport.Noop(),
			},
			Error: "Runtime field must be set before calling Start()",
		},
		{
			Name: "MissingTransport",
			Agent: agent.Agent{
				Log:     logr.Discard(),
				ID:      "1234",
				Runtime: runtime.Noop(),
			},
			Error: "Transport field must be set before calling Start()",
		},
		{
			Name: "InitializedCorrectly",
			Agent: agent.Agent{
				Log:       logr.Discard(),
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

func TestAgent_ConcurrentWorkflows(t *testing.T) {
	// The goal of this test is to ensure the agent rejects concurrent workflows. We have to
	// build a valid agent because it will also reject calls to HandleWorkflow without first
	// starting the Agent.

	logger := zapr.NewLogger(zap.Must(zap.NewDevelopment()))

	recorder := event.NoopRecorder()

	wrkflow := workflow.Workflow{
		ID: "1234",
		Actions: []workflow.Action{
			{
				ID:    "1234",
				Name:  "name",
				Image: "image",
			},
		},
	}

	trnport := agent.TransportMock{
		StartFunc: func(ctx context.Context, agentID string, handler workflow.Handler) error {
			return handler.HandleWorkflow(ctx, wrkflow, recorder)
		},
	}

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

	// Build a cancellable context so we can tear the goroutine down.

	errs := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { errs <- agnt.Start(ctx) }()

	// Await either an error or the mock runtime to tell us its stated.
	select {
	case err := <-errs:
		t.Fatalf("Unexpected error: %v", err)
	case <-started:
	}

	// Attempt to fire off another workflow.
	err := agnt.HandleWorkflow(context.Background(), wrkflow, recorder)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

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
		ID:      wrkflow.ID,
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
			recorder := event.NoopRecorder()

			trnport := agent.TransportMock{
				StartFunc: func(ctx context.Context, agentID string, handler workflow.Handler) error {
					return handler.HandleWorkflow(ctx, tc.Workflow, recorder)
				},
			}

			rntime := agent.ContainerRuntimeMock{
				RunFunc: func(ctx context.Context, action workflow.Action) error {
					if res, ok := tc.Errors[action.ID]; ok {
						return failure.NewReason(res.Message, res.Reason)
					}
					return nil
				},
			}

			agnt := agent.Agent{
				Log:       logger,
				Transport: &trnport,
				Runtime:   &rntime,
				ID:        "1234",
			}

			err := agnt.Start(context.Background())
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

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
