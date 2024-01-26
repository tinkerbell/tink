package worker

import (
	"context"
	"path"
	"path/filepath"
	"regexp"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/internal/proto"
)

const (
	errFailedToWait   = "failed to wait for completion of action"
	errFailedToRunCmd = "failed to run on-timeout command"

	infoWaitFinished = "wait finished for failed or timeout container"
)

// DockerClient is a subset of the interfaces implemented by docker's client.Client.
type DockerClient interface {
	client.ImageAPIClient
	client.ContainerAPIClient
}

type containerManager struct {
	logger          logr.Logger
	cli             DockerClient
	registryDetails RegistryConnDetails
}

// getLogger is a helper function to get logging out of a context, or use the default logger.
func (m *containerManager) getLogger(ctx context.Context) logr.Logger {
	loggerIface := ctx.Value(loggingContextKey)
	if loggerIface == nil {
		return m.logger
	}
	l, _ := loggerIface.(logr.Logger)
	return l
}

// NewContainerManager returns a new container manager.
func NewContainerManager(logger logr.Logger, cli DockerClient, registryDetails RegistryConnDetails) ContainerManager {
	return &containerManager{logger, cli, registryDetails}
}

func (m *containerManager) CreateContainer(ctx context.Context, cmd []string, wfID string, action *proto.WorkflowAction, captureLogs, privileged bool) (string, error) {
	l := m.getLogger(ctx)
	config := &container.Config{
		Image:        path.Join(m.registryDetails.Registry, action.GetImage()),
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		Tty:          true,
		Env:          action.GetEnvironment(),
	}
	if !captureLogs {
		config.AttachStdout = false
		config.AttachStderr = false
		config.Tty = false
	}

	wfDir := filepath.Join(defaultDataDir, wfID)
	hostConfig := &container.HostConfig{
		Privileged: privileged,
		Binds:      []string{wfDir + ":/workflow"},
	}

	if pidConfig := action.GetPid(); pidConfig != "" {
		hostConfig.PidMode = container.PidMode(pidConfig)
	}

	hostConfig.Binds = append(hostConfig.Binds, action.GetVolumes()...)
	l.Info("creating container", "command", cmd)
	name := makeValidContainerName(action.GetName())
	resp, err := m.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}
	return resp.ID, nil
}

// makeValidContainerName returns a valid container name for docker.
// only [a-zA-Z0-9][a-zA-Z0-9_.-] are allowed.
func makeValidContainerName(name string) string {
	regex := regexp.MustCompile(`[^a-zA-Z0-9_.-]`)
	result := "action_" // so we don't need to regex on the first character different from the rest.
	return result + regex.ReplaceAllString(name, "_")
}

func (m *containerManager) StartContainer(ctx context.Context, id string) error {
	m.getLogger(ctx).Info("starting container", "containerID", id)
	return errors.Wrap(m.cli.ContainerStart(ctx, id, container.StartOptions{}), "DOCKER START")
}

func (m *containerManager) WaitForContainer(ctx context.Context, id string) (proto.State, error) {
	// Inspect whether the container is in running state
	if _, err := m.cli.ContainerInspect(ctx, id); err != nil {
		return proto.State_STATE_FAILED, nil //nolint:nilerr // error is not nil, but it returns nil
	}

	// send API call to wait for the container completion
	wait, errC := m.cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		if status.StatusCode == 0 {
			return proto.State_STATE_SUCCESS, nil
		}
		return proto.State_STATE_FAILED, nil
	case err := <-errC:
		return proto.State_STATE_FAILED, err
	case <-ctx.Done():
		return proto.State_STATE_TIMEOUT, ctx.Err()
	}
}

func (m *containerManager) WaitForFailedContainer(ctx context.Context, id string, failedActionStatus chan proto.State) {
	l := m.getLogger(ctx)
	// send API call to wait for the container completion
	wait, errC := m.cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		if status.StatusCode == 0 {
			failedActionStatus <- proto.State_STATE_SUCCESS
			return
		}
		failedActionStatus <- proto.State_STATE_FAILED
	case err := <-errC:
		l.Error(err, "")
		failedActionStatus <- proto.State_STATE_FAILED
	case <-ctx.Done():
		l.Error(ctx.Err(), "")
		failedActionStatus <- proto.State_STATE_TIMEOUT
	}
}

func (m *containerManager) RemoveContainer(ctx context.Context, id string) error {
	// create options for removing container
	opts := container.RemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}
	m.getLogger(ctx).Info("removing container", "containerID", id)

	// send API call to remove the container
	return errors.Wrap(m.cli.ContainerRemove(ctx, id, opts), "DOCKER STOP")
}
