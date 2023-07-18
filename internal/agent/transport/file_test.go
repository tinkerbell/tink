package transport_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/zerologr"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog"
	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/transport"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

func TestFile(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	expect := workflow.Workflow{
		ID: "test-workflow-id",
		Actions: []workflow.Action{
			{
				ID:               "test-action-1",
				Name:             "my test action",
				Image:            "docker.io/hub/alpine",
				Cmd:              "sh -c",
				Args:             []string{"echo", "action 1"},
				Env:              map[string]string{"foo": "bar"},
				Volumes:          []string{"mount:/foo/bar:ro"},
				NetworkNamespace: "custom-namespace",
			},
			{
				ID:               "test-action-2",
				Name:             "my test action",
				Image:            "docker.io/hub/alpine",
				Cmd:              "sh -c",
				Args:             []string{"echo", "action 2"},
				Env:              map[string]string{"foo": "bar"},
				Volumes:          []string{"mount:/foo/bar:ro"},
				NetworkNamespace: "custom-namespace",
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	handler := &transport.WorkflowHandlerMock{
		HandleWorkflowFunc: func(contextMoqParam context.Context, workflow workflow.Workflow, recorder event.Recorder) {
			if !cmp.Equal(expect, workflow) {
				t.Fatalf("Workflow diff:\n%v", cmp.Diff(expect, workflow))
			}
		},
	}

	f := transport.File{
		Log:  zerologr.New(&logger),
		Path: "./testdata/workflow.yml",
	}

	err := f.Start(ctx, "agent_id", handler)
	if err != nil {
		t.Fatal(err)
	}
}
