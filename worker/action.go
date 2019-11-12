package main

import (
	"bufio"
	"bytes"
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
	"github.com/pkg/errors"
)

var (
	registry string
	cli      *client.Client
)

func executeAction(ctx context.Context, action *pb.WorkflowAction) (string, int, error) {
	err := pullActionImage(ctx, action)
	if err != nil {
		return fmt.Sprintf("Failed to pull Image"), 1, errors.Wrap(err, "DOCKER PULL")
	}

	startedAt := time.Now()
	id, err := createContainer(ctx, action)
	if err != nil {
		return fmt.Sprintf("Failed to run container"), 1, errors.Wrap(err, "DOCKER CREATE")
	}

	stopLogs := make(chan bool)
	logs := new(bytes.Buffer)
	go func(srt time.Time, exit chan bool) {
		req := func(sr string) {
			// get logs the runtime container
			rc, err := getLogs(ctx, cli, id, strconv.FormatInt(srt.Unix(), 10))
			if err != nil {
				stopLogs <- true
			}
			defer rc.Close()

			// create new scanner from the container output
			scanner := bufio.NewScanner(rc)

			// scan entire container output
			for scanner.Scan() {
				// write all the logs from the scanner
				logs.Write(append(scanner.Bytes(), []byte("\n")...))
				io.Copy(os.Stdout, logs)
				// flush the buffer of logs
				logs.Reset()
			}
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
				time.Sleep(time.Millisecond * 500)
			}
		}
	}(startedAt, stopLogs)

	status, err := waitContainer(ctx, id, stopLogs)
	if err != nil {
		return fmt.Sprintf("Failed to wait for completion"), status, errors.Wrap(err, "DOCKER_WAIT")
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

func createContainer(ctx context.Context, action *pb.WorkflowAction) (string, error) {
	config := &container.Config{
		Image:        registry + "/" + action.GetImage(),
		AttachStdout: true,
		AttachStderr: true,
	}

	if action.Command != "" {
		config.Cmd = []string{action.Command}
	}

	resp, err := cli.ContainerCreate(ctx, config, nil, nil, action.GetName())
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", errors.Wrap(err, "DOCKER START")
	}
	return resp.ID, nil
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

func waitContainer(ctx context.Context, id string, stopLogs chan bool) (int, error) {
	// send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case status := <-wait:
		stopLogs <- true
		return int(status.StatusCode), nil
	case err := <-errC:
		stopLogs <- true
		return 0, err
	}
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
