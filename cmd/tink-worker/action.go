package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

var (
	registry string
	cli      *client.Client
)

const (
	errCreateContainer = "failed to create container"
	errRemoveContainer = "failed to remove container"
	errFailedToWait    = "failed to wait for completion of action"
	errFailedToRunCmd  = "failed to run on-timeout command"

	infoWaitFinished = "wait finished for failed or timeout container"
)

func executeAction(ctx context.Context, action *pb.WorkflowAction, wfID string) (pb.ActionState, error) {
	l := logger.With("workflowID", wfID, "workerID", action.GetWorkerId(), "actionName", action.GetName(), "actionImage", action.GetImage())
	err := pullActionImage(ctx, action)
	if err != nil {
		return pb.ActionState_ACTION_IN_PROGRESS, errors.Wrap(err, "DOCKER PULL")
	}
	id, err := createContainer(ctx, l, action, action.Command, wfID)
	if err != nil {
		return pb.ActionState_ACTION_IN_PROGRESS, errors.Wrap(err, "DOCKER CREATE")
	}
	l.With("containerID", id, "command", action.GetOnTimeout()).Info("container created")
	// Setting time context for action
	timeCtx := ctx
	if action.Timeout > 0 {
		var cancel context.CancelFunc
		timeCtx, cancel = context.WithTimeout(context.Background(), time.Duration(action.Timeout)*time.Second)
		defer cancel()
	}
	err = startContainer(timeCtx, l, id)
	if err != nil {
		return pb.ActionState_ACTION_IN_PROGRESS, errors.Wrap(err, "DOCKER RUN")
	}

	failedActionStatus := make(chan pb.ActionState)

	//capturing logs of action container in a go-routine
	go captureLogs(ctx, id)

	status, err := waitContainer(timeCtx, id)
	if err != nil {
		rerr := removeContainer(ctx, l, id)
		if rerr != nil {
			rerr = errors.Wrap(rerr, errRemoveContainer)
			l.With("containerID", id).Error(rerr)
			return status, rerr
		}
		return status, errors.Wrap(err, "DOCKER_WAIT")
	}
	rerr := removeContainer(ctx, l, id)
	if rerr != nil {
		return status, errors.Wrap(rerr, "DOCKER_REMOVE")
	}
	l.With("status", status.String()).Info("container removed")
	if status != pb.ActionState_ACTION_SUCCESS {
		if status == pb.ActionState_ACTION_TIMEOUT && action.OnTimeout != nil {
			id, err = createContainer(ctx, l, action, action.OnTimeout, wfID)
			if err != nil {
				l.Error(errors.Wrap(err, errCreateContainer))
			}
			l.With("containerID", id, "status", status.String(), "command", action.GetOnTimeout()).Info("container created")
			failedActionStatus := make(chan pb.ActionState)
			go captureLogs(ctx, id)
			go waitFailedContainer(ctx, id, failedActionStatus)
			err = startContainer(ctx, l, id)
			if err != nil {
				l.Error(errors.Wrap(err, errFailedToRunCmd))
			}
			onTimeoutStatus := <-failedActionStatus
			l.With("status", onTimeoutStatus).Info("action timeout")
		} else {
			if action.OnFailure != nil {
				id, err = createContainer(ctx, l, action, action.OnFailure, wfID)
				if err != nil {
					l.Error(errors.Wrap(err, errFailedToRunCmd))
				}
				l.With("containerID", id, "actionStatus", status.String(), "command", action.GetOnFailure()).Info("container created")
				go captureLogs(ctx, id)
				go waitFailedContainer(ctx, id, failedActionStatus)
				err = startContainer(ctx, l, id)
				if err != nil {
					l.Error(errors.Wrap(err, errFailedToRunCmd))
				}
				onFailureStatus := <-failedActionStatus
				l.With("status", onFailureStatus).Info("action failed")
			}
		}
		l.Info(infoWaitFinished)
		if err != nil {
			rerr := removeContainer(ctx, l, id)
			if rerr != nil {
				l.Error(errors.Wrap(rerr, errRemoveContainer))
			}
			l.Error(errors.Wrap(err, errFailedToWait))
		}
		rerr = removeContainer(ctx, l, id)
		if rerr != nil {
			l.Error(errors.Wrap(rerr, errRemoveContainer))
		}
	}
	l.With("status", status).Info("action container exited")
	return status, nil
}

func captureLogs(ctx context.Context, id string) {
	reader, err := cli.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func pullActionImage(ctx context.Context, action *pb.WorkflowAction) error {
	user := os.Getenv("REGISTRY_USERNAME")
	pwd := os.Getenv("REGISTRY_PASSWORD")
	if user == "" || pwd == "" {
		return errors.New("required REGISTRY_USERNAME and REGISTRY_PASSWORD")
	}

	authConfig := types.AuthConfig{
		Username:      user,
		Password:      pwd,
		ServerAddress: registry,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return errors.Wrap(err, "DOCKER AUTH")
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	out, err := cli.ImagePull(ctx, registry+"/"+action.GetImage(), types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return errors.Wrap(err, "DOCKER PULL")
	}
	defer out.Close()
	if _, err := io.Copy(os.Stdout, out); err != nil {
		return err
	}
	return nil
}

func createContainer(ctx context.Context, l log.Logger, action *pb.WorkflowAction, cmd []string, wfID string) (string, error) {
	config := &container.Config{
		Image:        registry + "/" + action.GetImage(),
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Env:          action.GetEnvironment(),
	}
	if cmd != nil {
		config.Cmd = cmd
	}

	wfDir := dataDir + string(os.PathSeparator) + wfID
	hostConfig := &container.HostConfig{
		Privileged: true,
		Binds:      []string{wfDir + ":/workflow"},
	}
	hostConfig.Binds = append(hostConfig.Binds, action.GetVolumes()...)
	l.With("command", cmd).Info("creating container")
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, action.GetName())
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}
	return resp.ID, nil
}

func startContainer(ctx context.Context, l log.Logger, id string) error {
	l.With("containerID", id).Debug("starting container")
	err := cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "DOCKER START")
	}
	return nil
}

func waitContainer(ctx context.Context, id string) (pb.ActionState, error) {
	// Inspect whether the container is in running state
	_, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return pb.ActionState_ACTION_FAILED, nil
	}

	// send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		if status.StatusCode == 0 {
			return pb.ActionState_ACTION_SUCCESS, nil
		}
		return pb.ActionState_ACTION_FAILED, nil
	case err := <-errC:
		return pb.ActionState_ACTION_FAILED, err
	case <-ctx.Done():
		return pb.ActionState_ACTION_TIMEOUT, ctx.Err()
	}
}

func waitFailedContainer(ctx context.Context, id string, failedActionStatus chan pb.ActionState) {
	// send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		if status.StatusCode == 0 {
			failedActionStatus <- pb.ActionState_ACTION_SUCCESS
		}
		failedActionStatus <- pb.ActionState_ACTION_FAILED
	case err := <-errC:
		logger.Error(err)
		failedActionStatus <- pb.ActionState_ACTION_FAILED
	}
}

func removeContainer(ctx context.Context, l log.Logger, id string) error {
	// create options for removing container
	opts := types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}
	l.With("containerID", id).Info("removing container")

	// send API call to remove the container
	err := cli.ContainerRemove(ctx, id, opts)
	if err != nil {
		return err
	}
	return nil
}

func initializeDockerClient() (*client.Client, error) {
	registry = os.Getenv("DOCKER_REGISTRY")
	if registry == "" {
		return nil, errors.New("required DOCKER_REGISTRY")
	}
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "DOCKER CLIENT")
	}
	return c, nil
}
