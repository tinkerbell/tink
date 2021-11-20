package cmd

import (
	"context"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type containerAction struct {
	actionPid      string
	actionVolumes  string
	actionEnv      []string
	actionName     string
	volumeBindings []string
}

var containerState = map[int32]string{
	0: "SUCCESS",
	1: "FAILED",
	2: "TIMEOUT",
}

// createContainer creates container and returns containerid on success else returns error.
func createContainer(ctx context.Context, cli client.ContainerAPIClient, cActions containerAction, cmd []string, registry string, imageName string, captureLogs bool) (string, error) {
	config := &container.Config{
		Image:        path.Join(registry, imageName),
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		Tty:          true,
		Env:          cActions.actionEnv,
	}
	if !captureLogs {
		config.AttachStdout = true
		config.AttachStderr = true
		config.Tty = false
	}

	hostConfig := &container.HostConfig{
		Privileged: true,
		Binds:      cActions.volumeBindings,
	}

	if pidConfig := cActions.actionPid; pidConfig != "" {
		hostConfig.PidMode = container.PidMode(pidConfig)
	}

	hostConfig.Binds = append(hostConfig.Binds, cActions.actionVolumes)
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, cActions.actionName)
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}
	return resp.ID, nil
}

// startContainer starts container and returns error on failure.
func startContainer(ctx context.Context, cli client.ContainerAPIClient, containerId string) error {
	err := cli.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "DOCKER START")
	}
	return nil
}

// removeContainer delete container and remove volumes, links if flags are set to true. returns error on failure.
func removeContainer(ctx context.Context, cli client.ContainerAPIClient, containerId string, forceRemoveContainer bool, linkRemove bool, volumesRemove bool) error {
	// create options for removing container
	opts := types.ContainerRemoveOptions{
		Force:         forceRemoveContainer,
		RemoveLinks:   linkRemove,
		RemoveVolumes: volumesRemove,
	}

	// send API call to remove the container
	err := cli.ContainerRemove(ctx, containerId, opts)
	if err != nil {
		return errors.Wrap(err, "DOCKER REMOVE")
	}
	return nil
}

//waitContainer waits on any non "not-running" states and returns container state on success or returns error on failure.
func waitContainer(ctx context.Context, cli client.ContainerAPIClient, containerId string) (string, error) {

	// Send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case status := <-wait:
		if status.StatusCode == 0 {
			return containerState[0], nil
		}
		return containerState[1], nil

	case err := <-errC:
		return containerState[1], err

	case <-ctx.Done():
		return containerState[2], ctx.Err()
	}
}
