package container

import (
	"context"
	"path"
	"reflect"
	"regexp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type Cdata struct {
	Data ConfigData
}

type ConfigData struct {
	Name string
	Pid  string
	Env  []string
	// VolumeBindings format: /dev:/dev .
	VolumeBindings []string
	Cmd            []string
	ImageName      string
	Registry       string
	CaptureLogs    bool
}

type RemoveOpts struct {
	Force   bool
	Links   bool
	Volumes bool
}

// Create creates container and returns containerid on success else returns error.
func (c Cdata) Create(ctx context.Context, cli client.ContainerAPIClient) (string, error) {
	if c.Data.Name == "" || c.Data.ImageName == "" {
		return "", errors.New("container name and image name are required data")
	}

	if cli == nil || (reflect.ValueOf(cli).IsNil()) {
		return "", errors.New("ContainerAPIClient interface is not valid")
	}

	if len(c.Data.Cmd) == 0 {
		return "", errors.New("cmd has invalid data")
	}

	if len(c.Data.VolumeBindings) > 0 {
		isValidVolume := regexp.MustCompile(`^[a-zA-Z0-9/:-]*$`).MatchString
		for _, value := range c.Data.VolumeBindings {
			if !isValidVolume(value) {
				return "", errors.New("volume is invalid format. valid format:'/dev:/dev'")
			}
		}
	}

	config := &container.Config{
		Image:        path.Join(c.Data.Registry, c.Data.ImageName),
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          c.Data.Cmd,
		Tty:          true,
		Env:          c.Data.Env,
	}
	if !c.Data.CaptureLogs {
		config.AttachStdout = false
		config.AttachStderr = false
		config.Tty = false
	}

	hostConfig := &container.HostConfig{
		Privileged: true,
		Binds:      c.Data.VolumeBindings,
	}

	if pidConfig := c.Data.Pid; pidConfig != "" {
		hostConfig.PidMode = container.PidMode(pidConfig)
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, c.Data.Name)
	if err != nil {
		return "", errors.Wrap(err, "container create failed")
	}
	return resp.ID, nil
}

// Start starts container and returns error on failure.
func (c Cdata) Start(ctx context.Context, cli client.ContainerAPIClient, id string) error {
	if id == "" {
		return errors.New("empty string is not a valid id")
	}

	if cli == nil || (reflect.ValueOf(cli).IsNil()) {
		return errors.New("ContainerAPIClient interface is not valid")
	}

	err := cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "container start failed")
	}
	return nil
}

// Remove delete container and remove volumes, links if flags are set to true. returns error on failure.
func (c Cdata) Remove(ctx context.Context, cli client.ContainerAPIClient, id string, opts RemoveOpts) error {
	if id == "" {
		return errors.New("empty string is not a valid id")
	}

	if cli == nil || (reflect.ValueOf(cli).IsNil()) {
		return errors.New("ContainerAPIClient interface is not valid")
	}

	// create options for removing container
	options := types.ContainerRemoveOptions{
		Force:         opts.Force,
		RemoveLinks:   opts.Links,
		RemoveVolumes: opts.Volumes,
	}

	// send API call to remove the container
	err := cli.ContainerRemove(ctx, id, options)
	if err != nil {
		return errors.Wrap(err, "container remove failed")
	}
	return nil
}

// Wait waits on any "non-running" container states and returns "SUCCESS" if status code is 0, returns FAILED on err , returns TIMEOUT on failed and end of communication.
// WaitConditionNotRunning is used to wait for any of the non-running states: "created", "exited", "dead", "removing", or "removed".
func (c Cdata) Wait(ctx context.Context, cli client.ContainerAPIClient, id string) (string, error) {
	if id == "" {
		return "", errors.New("empty string is not a valid id")
	}

	if cli == nil || (reflect.ValueOf(cli).IsNil()) {
		return "", errors.New("ContainerAPIClient interface is not valid")
	}

	containerState := map[int32]string{
		0: "SUCCESS",
		1: "FAILED",
		2: "TIMEOUT",
	}

	if _, err := cli.ContainerInspect(ctx, id); err != nil {
		return containerState[1], err
	}

	// Send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
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
