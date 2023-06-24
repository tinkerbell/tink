package transport

import (
	"context"
	"errors"
	"io"

	"github.com/avast/retry-go"
	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/workflow"
	workflowproto "github.com/tinkerbell/tink/internal/proto/workflow/v2"
)

var _ event.Recorder = &GRPC{}

func NewGRPC(log logr.Logger, client workflowproto.WorkflowServiceClient) *GRPC {
	return &GRPC{
		log:    log,
		client: client,
	}
}

type GRPC struct {
	log    logr.Logger
	client workflowproto.WorkflowServiceClient
}

func (g *GRPC) Start(ctx context.Context, agentID string, handler WorkflowHandler) error {
	stream, err := g.client.GetWorkflows(ctx, &workflowproto.GetWorkflowsRequest{
		AgentId: agentID,
	})
	if err != nil {
		return err
	}

	for {
		request, err := stream.Recv()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		}

		switch request.GetCmd().(type) {
		case *workflowproto.GetWorkflowsResponse_StartWorkflow_:
			grpcWorkflow := request.GetStartWorkflow().GetWorkflow()

			if err := validateGRPCWorkflow(grpcWorkflow); err != nil {
				g.log.Info(
					"Dropping request to start workflow; invalid payload",
					"error", err,
					"payload", grpcWorkflow,
				)
				continue
			}

			handler.HandleWorkflow(ctx, toWorkflow(grpcWorkflow), g)

		case *workflowproto.GetWorkflowsResponse_StopWorkflow_:
			if request.GetStopWorkflow().WorkflowId == "" {
				g.log.Info("Dropping request to cancel workflow; missing workflow ID")
				continue
			}

			handler.CancelWorkflow(request.GetStopWorkflow().WorkflowId)
		}
	}
}

func (g *GRPC) RecordEvent(ctx context.Context, e event.Event) error {
	evnt, err := toGRPC(e)
	if err != nil {
		return err
	}

	publish := func() error {
		payload := workflowproto.PublishEventRequest{Event: evnt}
		if _, err := g.client.PublishEvent(ctx, &payload); err != nil {
			return err
		}
		return nil
	}

	return retry.Do(publish, retry.Attempts(5), retry.DelayType(retry.BackOffDelay))
}

func validateGRPCWorkflow(wflw *workflowproto.Workflow) error {
	if wflw == nil {
		return errors.New("workflow must not be nil")
	}

	for _, action := range wflw.Actions {
		if action == nil {
			return errors.New("workflow actions must not be nil")
		}
	}

	return nil
}

func toWorkflow(wflw *workflowproto.Workflow) workflow.Workflow {
	return workflow.Workflow{
		ID:      wflw.WorkflowId,
		Actions: toActions(wflw.GetActions()),
	}
}

func toActions(a []*workflowproto.Workflow_Action) []workflow.Action {
	var actions []workflow.Action
	for _, action := range a {
		actions = append(actions, workflow.Action{
			ID:               action.GetId(),
			Name:             action.GetName(),
			Image:            action.GetImage(),
			Cmd:              action.GetCmd(),
			Args:             action.GetArgs(),
			Env:              action.GetEnv(),
			Volumes:          action.GetVolumes(),
			NetworkNamespace: action.GetNetworkNamespace(),
		})
	}
	return actions
}

func toGRPC(e event.Event) (*workflowproto.Event, error) {
	switch v := e.(type) {
	case event.ActionStarted:
		return &workflowproto.Event{
			WorkflowId: v.WorkflowID,
			Event: &workflowproto.Event_ActionStarted_{
				ActionStarted: &workflowproto.Event_ActionStarted{
					ActionId: v.ActionID,
				},
			},
		}, nil
	case event.ActionSucceeded:
		return &workflowproto.Event{
			WorkflowId: v.WorkflowID,
			Event: &workflowproto.Event_ActionSucceeded_{
				ActionSucceeded: &workflowproto.Event_ActionSucceeded{
					ActionId: v.ActionID,
				},
			},
		}, nil
	case event.ActionFailed:
		return &workflowproto.Event{
			WorkflowId: v.WorkflowID,
			Event: &workflowproto.Event_ActionFailed_{
				ActionFailed: &workflowproto.Event_ActionFailed{
					ActionId:       v.ActionID,
					FailureReason:  &v.Reason,
					FailureMessage: &v.Message,
				},
			},
		}, nil
	case event.WorkflowRejected:
		return &workflowproto.Event{
			WorkflowId: v.ID,
			Event: &workflowproto.Event_WorkflowRejected_{
				WorkflowRejected: &workflowproto.Event_WorkflowRejected{
					Message: v.Message,
				},
			},
		}, nil
	}

	return nil, event.IncompatibleError{Event: e}
}
