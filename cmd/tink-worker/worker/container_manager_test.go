package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/go-logr/zapr"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/tinkerbell/tink/internal/proto"
	"go.uber.org/zap"
)

type fakeDockerClient struct {
	client.ImageAPIClient
	client.ContainerAPIClient

	imagePullContent string
	containerID      string
	delay            time.Duration
	statusCode       int
	err              error
	waitErr          error
	imageInspectErr  error
}

type dockerClientOpt func(*fakeDockerClient)

func withImageInspectErr(err error) dockerClientOpt {
	return func(c *fakeDockerClient) {
		c.imageInspectErr = err
	}
}

func newFakeDockerClient(containerID, imagePullContent string, delay time.Duration, statusCode int, err, waitErr error, opts ...dockerClientOpt) *fakeDockerClient {
	f := &fakeDockerClient{
		containerID:      containerID,
		imagePullContent: imagePullContent,
		delay:            delay,
		statusCode:       statusCode,
		err:              err,
		waitErr:          waitErr,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (c *fakeDockerClient) ContainerCreate(
	context.Context, *container.Config, *container.HostConfig, *network.NetworkingConfig, *specs.Platform, string,
) (container.CreateResponse, error) {
	if c.err != nil {
		return container.CreateResponse{}, c.err
	}

	return container.CreateResponse{
		ID: c.containerID,
	}, nil
}

func (c *fakeDockerClient) ContainerStart(context.Context, string, container.StartOptions) error {
	if c.err != nil {
		return c.err
	}
	return nil
}

func (c *fakeDockerClient) ContainerInspect(context.Context, string) (dockertypes.ContainerJSON, error) {
	if c.err != nil {
		return dockertypes.ContainerJSON{}, c.err
	}
	return dockertypes.ContainerJSON{}, nil
}

func (c *fakeDockerClient) ContainerWait(context.Context, string, container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	respChan := make(chan container.WaitResponse)
	errChan := make(chan error)
	go func(e error) {
		time.Sleep(c.delay)
		if e != nil {
			errChan <- e
			return
		}
		respChan <- container.WaitResponse{
			StatusCode: int64(c.statusCode),
		}
	}(c.waitErr)
	return respChan, errChan
}

func (c *fakeDockerClient) ContainerRemove(context.Context, string, container.RemoveOptions) error {
	if c.err != nil {
		return c.err
	}
	return nil
}

func TestContainerManagerCreate(t *testing.T) {
	cases := []struct {
		name         string
		workflowName string
		action       *proto.WorkflowAction
		containerID  string
		registry     string
		clientErr    error
		wantErr      error
	}{
		{
			name:         "Happy Path",
			workflowName: "saveTheRebelBase",
			action: &proto.WorkflowAction{
				TaskName:    "UseTheForce",
				Name:        "blow up the death star",
				Image:       "yav.in/4/forestmoon",
				Environment: []string{"MODE=insane", ""},
				Volumes:     []string{"/tie-fighter/darth_vader:/behind_you"},
				Pid:         "1",
			},
			containerID: "nomedalforchewie",
			registry:    "rebelba.se",
		},
		{
			name:         "create failure",
			workflowName: "saveTheRebelBase",
			action: &proto.WorkflowAction{
				TaskName:    "UseTheForce",
				Name:        "blow up the death star",
				Image:       "yav.in/4/forestmoon",
				Environment: []string{"MODE=insane", ""},
				Volumes:     []string{"/tie-fighter/darth_vader:/behind_you"},
				Pid:         "1",
			},
			containerID: "nomedalforchewie",
			registry:    "rebelba.se",
			clientErr:   errors.New("You missed the shot"),
			wantErr:     errors.New("DOCKER CREATE: You missed the shot"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zapr.NewLogger(zap.Must(zap.NewDevelopment()))
			mgr := NewContainerManager(logger, newFakeDockerClient(tc.containerID, "", 0, 0, tc.clientErr, nil), RegistryConnDetails{Registry: tc.registry})

			ctx := context.Background()
			got, gotErr := mgr.CreateContainer(ctx, []string{}, tc.workflowName, tc.action, false, true)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}

			if got != tc.containerID {
				t.Errorf("Unexpected response: got '%s', expected '%s'", got, tc.containerID)
			}
		})
	}
}

func TestContainerManagerStart(t *testing.T) {
	cases := []struct {
		name        string
		containerID string
		clientErr   error
		wantErr     error
	}{
		{
			name:        "Happy Path",
			containerID: "nomedalforchewie",
		},
		{
			name:        "start failure",
			containerID: "nomedalforchewie",
			clientErr:   errors.New("You missed the shot"),
			wantErr:     errors.New("DOCKER START: You missed the shot"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zapr.NewLogger(zap.Must(zap.NewDevelopment()))
			mgr := NewContainerManager(logger, newFakeDockerClient(tc.containerID, "", 0, 0, tc.clientErr, nil), RegistryConnDetails{Registry: ""})

			ctx := context.Background()
			gotErr := mgr.StartContainer(ctx, tc.containerID)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}
		})
	}
}

func TestContainerManagerWait(t *testing.T) {
	cases := []struct {
		name           string
		containerID    string
		dockerResponse int
		contextTimeout time.Duration
		clientErr      error
		waitErr        error
		wantState      proto.State
		wantErr        error
	}{
		{
			name:           "Happy Path",
			containerID:    "nomedalforchewie",
			dockerResponse: 0,
			wantState:      proto.State_STATE_SUCCESS,
		},
		{
			name:           "start failure",
			containerID:    "chewieDied",
			dockerResponse: 1,
			wantState:      proto.State_STATE_FAILED,
			waitErr:        nil,
		},
		{
			name:           "client wait failure",
			containerID:    "nomedalforchewie",
			dockerResponse: 1,
			wantState:      proto.State_STATE_FAILED,
			waitErr:        errors.New("Vader Won"),
			wantErr:        errors.New("Vader Won"),
		},
		{
			name:        "client inspect failure",
			containerID: "nomedalforchewie",
			wantState:   proto.State_STATE_FAILED,
			clientErr:   errors.New("inspect failed"),
			wantErr:     nil,
		},
		{
			name:           "client timeout",
			containerID:    "nomedalforchewie",
			wantState:      proto.State_STATE_TIMEOUT,
			contextTimeout: time.Millisecond * 2,
			waitErr:        errors.New("Vader Won"),
			wantErr:        errors.New("context deadline exceeded"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zapr.NewLogger(zap.Must(zap.NewDevelopment()))
			mgr := NewContainerManager(logger, newFakeDockerClient(tc.containerID, "", time.Millisecond*20, tc.dockerResponse, tc.clientErr, tc.waitErr), RegistryConnDetails{Registry: ""})
			ctx, cancel := context.WithTimeout(context.Background(), tc.contextTimeout)
			defer cancel()
			if tc.contextTimeout == 0 {
				ctx = context.Background()
			}

			got, gotErr := mgr.WaitForContainer(ctx, tc.containerID)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}
			if got.String() != tc.wantState.String() {
				t.Errorf("Unexpected response: got %s wanted %s", got, tc.wantState)
			}
		})
	}
}

func TestContainerManagerWaitFailed(t *testing.T) {
	cases := []struct {
		name           string
		containerID    string
		dockerResponse int
		contextTimeout time.Duration
		waitTime       time.Duration
		clientErr      error
		wantState      proto.State
	}{
		{
			name:           "Happy Path",
			containerID:    "nomedalforchewie",
			dockerResponse: 0,
			waitTime:       0,
			wantState:      proto.State_STATE_SUCCESS,
		},
		{
			name:           "start failure",
			containerID:    "chewieDied",
			dockerResponse: 1,
			wantState:      proto.State_STATE_FAILED,
			clientErr:      nil,
		},
		{
			name:           "client wait failure",
			containerID:    "nomedalforchewie",
			dockerResponse: 1,
			wantState:      proto.State_STATE_FAILED,
			clientErr:      errors.New("Vader Won"),
		},
		{
			name:           "client timeout",
			containerID:    "nomedalforchewie",
			wantState:      proto.State_STATE_TIMEOUT,
			waitTime:       time.Millisecond * 20,
			contextTimeout: time.Millisecond * 10,
			clientErr:      errors.New("Vader Won"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zapr.NewLogger(zap.Must(zap.NewDevelopment()))
			mgr := NewContainerManager(logger, newFakeDockerClient(tc.containerID, "", tc.waitTime, tc.dockerResponse, nil, tc.clientErr), RegistryConnDetails{Registry: ""})
			ctx, cancel := context.WithTimeout(context.Background(), tc.contextTimeout)
			defer cancel()
			if tc.contextTimeout == 0 {
				ctx = context.Background()
			}
			failedChan := make(chan proto.State)
			go mgr.WaitForFailedContainer(ctx, tc.containerID, failedChan)
			got := <-failedChan

			if got.String() != tc.wantState.String() {
				t.Errorf("Unexpected response: got %s wanted %s", got, tc.wantState)
			}
		})
	}
}

func TestContainerManagerRemove(t *testing.T) {
	cases := []struct {
		name        string
		containerID string
		clientErr   error
		wantErr     error
	}{
		{
			name:        "Happy Path",
			containerID: "nomedalforchewie",
		},
		{
			name:        "start failure",
			containerID: "nomedalforchewie",
			clientErr:   errors.New("You missed the shot"),
			wantErr:     errors.New("DOCKER STOP: You missed the shot"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zapr.NewLogger(zap.Must(zap.NewDevelopment()))
			mgr := NewContainerManager(logger, newFakeDockerClient(tc.containerID, "", 0, 0, tc.clientErr, nil), RegistryConnDetails{Registry: ""})

			ctx := context.Background()
			gotErr := mgr.RemoveContainer(ctx, tc.containerID)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}
		})
	}
}
