package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	pb "github.com/packethost/rover/protos/rover"
	workflowpb "github.com/packethost/rover/protos/workflow"
	"github.com/pkg/errors"
)

var (
	registry string
	cli      *client.Client
)

func executeAction(ctx context.Context, action *pb.WorkflowAction) (string, workflowpb.ActionState, error) {
	err := pullActionImage(ctx, action)
	if err != nil {
		return fmt.Sprintf("Failed to pull Image : %s", action.GetImage()), 1, errors.Wrap(err, "DOCKER PULL")
	}

	id, err := createContainer(ctx, action, action.Command)
	if err != nil {
		return fmt.Sprintf("Failed to create container"), 1, errors.Wrap(err, "DOCKER CREATE")
	}
	var timeCtx context.Context
	var cancel context.CancelFunc
	if action.Timeout > 0 {
		timeCtx, cancel = context.WithTimeout(context.Background(), time.Duration(action.Timeout)*time.Second)
	} else {
		timeCtx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
	}
	defer cancel()
	//run container with timeout context
	startedAt := time.Now()
	err = runContainer(timeCtx, id)
	if err != nil {
		return fmt.Sprintf("Failed to run container"), 1, errors.Wrap(err, "DOCKER RUN")
	}
	stopLogs := make(chan bool)
	go func(srt time.Time, exit chan bool) {
		req := func(sr string) {
			// get logs the runtime container
			rc, err := getLogs(ctx, cli, id, strconv.FormatInt(srt.Unix(), 10))
			if err != nil {
				stopLogs <- true
			}
			defer rc.Close()
			io.Copy(os.Stdout, rc)
		}
	Loop:
		for {
			select {
			case <-exit:
				// Last call to be sure to get the end of the logs content
				now := time.Now()
				now = now.Add(time.Second * -1)
				startAt := strconv.FormatInt(now.Unix(), 10)
				req(startAt)
				break Loop
			default:
				// Running call to trace the container logs every 500ms
				startAt := strconv.FormatInt(srt.Unix(), 10)
				srt = srt.Add(time.Millisecond * 500)
				req(startAt)
			}
		}
	}(startedAt, stopLogs)

	status, err := waitContainer(timeCtx, id, stopLogs)
	if err != nil {
		rerr := removeContainer(ctx, id)
		if rerr != nil {
			fmt.Println("Failed to remove container as ", rerr)
		}
		return fmt.Sprintf("Failed to wait for completion of action"), status, errors.Wrap(err, "DOCKER_WAIT")
	}
	rerr := removeContainer(ctx, id)
	if rerr != nil {
		return fmt.Sprintf("Failed to remove container of action"), status, errors.Wrap(rerr, "DOCKER_REMOVE")
	}
	if status != 0 {
		if status == workflowpb.ActionState_ACTION_FAILED && action.OnFailure != "" {
			id, err = createContainer(ctx, action, action.OnFailure)
			if err != nil {
				fmt.Println("Failed to create on-failure command: ", err)
			}
			err = runContainer(ctx, id)
			if err != nil {
				fmt.Println("Failed to run on-failure command: ", err)
			}
		} else if status == workflowpb.ActionState_ACTION_TIMEOUT && action.OnTimeout != "" {
			id, err = createContainer(ctx, action, action.OnTimeout)
			if err != nil {
				fmt.Println("Failed to create on-timeout command: ", err)
			}
			err = runContainer(ctx, id)
			if err != nil {
				fmt.Println("Failed to run on-timeout command: ", err)
			}
		}
		_, err = waitContainer(ctx, id, stopLogs)
		if err != nil {
			rerr := removeContainer(ctx, id)
			if rerr != nil {
				fmt.Println("Failed to remove container as ", rerr)
			}
			fmt.Println("Failed to wait for container : ", err)
		}
		rerr := removeContainer(ctx, id)
		if rerr != nil {
			fmt.Println("Failed to remove container as ", rerr)
		}
	}
	fmt.Println("Action container exits with status code ", status)
	return fmt.Sprintf("Successfull Execution"), status, nil
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
	io.Copy(os.Stdout, out)
	return nil
}

func createContainer(ctx context.Context, action *pb.WorkflowAction, cmd string) (string, error) {
	config := &container.Config{
		Image:        registry + "/" + action.GetImage(),
		AttachStdout: true,
		AttachStderr: true,
	}

	if cmd != "" {
		config.Cmd = []string{cmd}
	}

	resp, err := cli.ContainerCreate(ctx, config, nil, nil, action.GetName())
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}
	return resp.ID, nil
}

func runContainer(ctx context.Context, id string) error {
	err := cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "DOCKER START")
	}
	return nil
}

func getLogs(ctx context.Context, cli *client.Client, id string, srt string) (io.ReadCloser, error) {

	fmt.Println("Capturing logs for container : ", id)
	// create options for capturing container logs
	opts := types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Since:      srt,
	}

	// send API call to capture the container logs
	logs, err := cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func waitContainer(ctx context.Context, id string, stopLogs chan bool) (workflowpb.ActionState, error) {
	// send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case status := <-wait:
		stopLogs <- true
		return workflowpb.ActionState(status.StatusCode), nil
	case err := <-errC:
		stopLogs <- true
		return workflowpb.ActionState_ACTION_FAILED, err
	case <-ctx.Done():
		stopLogs <- true
		return workflowpb.ActionState_ACTION_TIMEOUT, ctx.Err()
	}
}

func removeContainer(ctx context.Context, id string) error {
	// create options for removing container
	opts := types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}
	fmt.Println("Start removing container ", id)
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
		return nil, errors.New("requried DOCKER_REGISTRY")
	}
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "DOCKER CLIENT")
	}
	return c, nil
}
