package worker

import (
	"context"
	"path"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
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

// NewContainerManager returns a new container manager.
func NewContainerManager(logger logr.Logger, cli DockerClient, registryDetails RegistryConnDetails) ContainerManager {
	return &containerManager{logger, cli, registryDetails}
}

func (m *containerManager) CreateContainer(ctx context.Context, cmd []string, wfID string, action *pb.WorkflowAction, captureLogs, privileged bool) (string, error) {
	l := m.logger
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
	resp, err := m.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, action.GetName())
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}
	return resp.ID, nil
}

func (m *containerManager) StartContainer(ctx context.Context, id string) error {
	m.logger.V(1).Info("starting container", "containerID", id)
	return errors.Wrap(m.cli.ContainerStart(ctx, id, types.ContainerStartOptions{}), "DOCKER START")
}

func (m *containerManager) WaitForContainer(ctx context.Context, id string) (pb.State, error) {
	// Inspect whether the container is in running state
	if _, err := m.cli.ContainerInspect(ctx, id); err != nil {
		return pb.State_STATE_FAILED, nil // nolint:nilerr // error is not nil, but it returns nil
	}

	// send API call to wait for the container completion
	wait, errC := m.cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		if status.StatusCode == 0 {
			return pb.State_STATE_SUCCESS, nil
		}
		return pb.State_STATE_FAILED, nil
	case err := <-errC:
		return pb.State_STATE_FAILED, err
	case <-ctx.Done():
		return pb.State_STATE_TIMEOUT, ctx.Err()
	}
}

func (m *containerManager) WaitForFailedContainer(ctx context.Context, id string, failedActionStatus chan pb.State) {
	l := m.logger
	// send API call to wait for the container completion
	wait, errC := m.cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		if status.StatusCode == 0 {
			failedActionStatus <- pb.State_STATE_SUCCESS
			return
		}
		failedActionStatus <- pb.State_STATE_FAILED
	case err := <-errC:
		l.Error(err, "")
		failedActionStatus <- pb.State_STATE_FAILED
	case <-ctx.Done():
		l.Error(ctx.Err(), "")
		failedActionStatus <- pb.State_STATE_TIMEOUT
	}
}

func (m *containerManager) RemoveContainer(ctx context.Context, id string) error {
	// create options for removing container
	opts := types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}
	m.logger.Info("removing container", "containerID", id)

	// send API call to remove the container
	return errors.Wrap(m.cli.ContainerRemove(ctx, id, opts), "DOCKER STOP")
}
