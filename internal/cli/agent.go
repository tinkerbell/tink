package cli

import (
	"context"
	"fmt"

	"github.com/go-logr/zapr"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tinkerbell/tink/internal/agent"
	"github.com/tinkerbell/tink/internal/agent/runtime"
	"github.com/tinkerbell/tink/internal/agent/transport"
	"github.com/tinkerbell/tink/internal/agent/workflow"
	"go.uber.org/zap"
)

func NewAgent() ffcli.Command {
	return ffcli.Command{
		Exec: execAgent,
	}
}

/*
	Create the gRPC Client
	On getting a workflow, mark self as executing workflow.
	Download images starting with the first action.
	For each action
		Issue started event
		Run action
		Issue complete event (failure or success)
	After last action mark self as available for workflow.

	How do we deal with accept/reject workflows?
*/

func execAgent(ctx context.Context, _ []string) error {
	agentID := "1850e035-64b3-4311-b258-8138a5c40105"
	logger := zapr.NewLogger(zap.Must(zap.NewProduction()))

	rt, err := runtime.NewDocker()
	if err != nil {
		return fmt.Errorf("create runtime: %w", err)
	}

	tpt := transport.Fake{
		Log: logger,
		Workflows: []workflow.Workflow{
			{
				ID: "e0e70707-7866-4cad-baf9-091c653e820c",
				Actions: []workflow.Action{
					{
						ID:    "5b79e2d8-7190-41ea-84eb-91528c159f0d",
						Name:  "foobar",
						Image: "alpine",
						Cmd:   "sh",
						Args: []string{
							"-c", "echo", "hello world",
						},
					},
				},
			},
		},
	}

	return (&agent.Agent{
		Log:       logger,
		ID:        agentID,
		Transport: tpt,
		Runtime:   rt,
	}).Start(ctx)
}
