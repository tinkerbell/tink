package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/tinkerbell/tink/internal/agent/failure"
	"github.com/tinkerbell/tink/internal/agent/runtime"
	"github.com/tinkerbell/tink/internal/agent/workflow"
	"go.uber.org/multierr"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func TestDockerImageNotPresent(t *testing.T) {
	clnt, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}

	image := "hello-world"

	images, err := clnt.ImageList(context.Background(), types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", image)),
	})
	if err != nil {
		t.Fatalf("Unexpected error listing images: %v", err)
	}

	var errSum error
	for _, image := range images {
		_, err := clnt.ImageRemove(context.Background(), image.ID, types.ImageRemoveOptions{})
		if err != nil {
			errSum = multierr.Append(errSum, fmt.Errorf("deleting image (%v): %v", image.ID, err))
		}
	}
	if errSum != nil {
		t.Error(err.Error())
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	rt, err := runtime.NewDocker(runtime.WithLogger(zerologr.New(&logger)))
	if err != nil {
		t.Fatal(err.Error())
	}

	action := workflow.Action{
		ID:    "foobar",
		Image: image,
	}

	err = rt.Run(context.Background(), action)
	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}
}

func TestDockerFailureMessages(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})

	for _, tc := range []struct {
		Name          string
		Action        workflow.Action
		ExpectReason  string
		ExpectMessage string
	}{
		{
			Name: "FailureReasonAndMessage",
			Action: workflow.Action{
				ID:    "foobar",
				Image: "alpine",
				Args:  command{Message: "failure message", Reason: "FailureReason", Code: 1}.Build(),
			},
			ExpectReason:  "FailureReason",
			ExpectMessage: "failure message",
		},
		{
			Name: "FailureMessageWithoutReason",
			Action: workflow.Action{
				ID:    "foobar",
				Image: "alpine",
				Args:  command{Message: "failure message", Code: 1}.Build(),
			},
			ExpectMessage: "failure message",
		},
		{
			Name: "NoError",
			Action: workflow.Action{
				ID:    "foobar",
				Image: "alpine",
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			rt, err := runtime.NewDocker(runtime.WithLogger(zerologr.New(&logger)))
			if err != nil {
				t.Fatal(err.Error())
			}

			err = rt.Run(context.Background(), tc.Action)
			reason, ok := failure.Reason(err)

			switch {
			case tc.ExpectMessage != "" && err == nil:
				t.Fatal("Expected error but received none")
			case tc.ExpectMessage != "" && tc.ExpectMessage != err.Error():
				t.Fatalf("Expected: %v; Received: %v", tc.ExpectMessage, err)
			case tc.ExpectMessage == "" && err != nil:
				t.Fatalf("Received unexpected error: %v", err)
			}

			switch {
			case tc.ExpectReason != "" && !ok:
				t.Fatal("Expected reason but found none")
			case tc.ExpectReason != "" && tc.ExpectReason != reason:
				t.Fatalf("Expected: %v; Received: %v", tc.ExpectReason, reason)
			case tc.ExpectReason == "" && ok:
				t.Fatalf("Received unexpected reason: %v", reason)
			}

			if tc.ExpectMessage == "" && tc.ExpectReason == "" && err != nil {
				t.Fatalf("Received unexpected error: %v", err)
			}
		})
	}
}

func TestDockerContextTimeout(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	rt, err := runtime.NewDocker(runtime.WithLogger(zerologr.New(&logger)))
	if err != nil {
		t.Fatal(err.Error())
	}

	action := workflow.Action{
		ID:    "foobar",
		Image: "alpine",
		Args:  command{Sleep: 30}.Build(),
	}

	// Give just enough time to launch the container and timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = rt.Run(ctx, action)

	if err == nil {
		t.Fatal("Expected error but received none")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expect: %v; Received: %v", context.DeadlineExceeded, err)
	}
}

type command struct {
	// Reason is the reason to write to /tinkerbell/failure-reason
	Reason string

	// Message is the message to write to /tinkerbell/failure-message
	Message string

	// Code is the exit code
	Code int

	// Sleep is the time to sleep in seconds before exiting.
	Sleep int
}

func (c command) Build() []string {
	cmds := []string{"trap exit SIGTERM"}

	if c.Reason != "" {
		cmds = append(cmds, fmt.Sprintf("echo %v >> /tinkerbell/failure-reason", c.Reason))
	}

	if c.Message != "" {
		cmds = append(cmds, fmt.Sprintf("echo %v >> /tinkerbell/failure-message", c.Message))
	}

	if c.Sleep > 0 {
		cmds = append(cmds, fmt.Sprintf("sleep %vs", c.Sleep))
	}

	cmds = append(cmds, fmt.Sprintf("exit %v", c.Code))

	return []string{"sh", "-c", strings.Join(cmds, "; ")}
}
