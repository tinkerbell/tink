package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"

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

func executeAction(ctx context.Context, action *pb.WorkflowAction) error {
	err := pullActionImage(ctx, action)
	if err != nil {
		return errors.Wrap(err, "DOCKER PULL")
	}

	id, err := createContainer(ctx, action)
	if err != nil {
		return errors.Wrap(err, "DOCKER CREATE")
	}

	out, err := getLogs(ctx, cli, id)
	defer out.Close()
	if err != nil {
		return errors.Wrap(err, "DOCKER LOGS")
	}
	io.Copy(os.Stdout, out)
	return nil
}

func pullActionImage(ctx context.Context, action *pb.WorkflowAction) error {
	user := os.Getenv("REGISTRY_USERNAME")
	pwd := os.Getenv("REGISTRY_PASSWORD")
	if user == "" || pwd == "" {
		return errors.New("requried REGISTRY_USERNAME and REGISTRY_PASSWORD")
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

	_, err = cli.ImagePull(ctx, registry+"/"+action.GetImage(), types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return errors.Wrap(err, "DOCKER PULL")
	}
	return nil
}

func createContainer(ctx context.Context, action *pb.WorkflowAction) (string, error) {
	config := &container.Config{
		Image:        registry + "/" + action.GetImage(),
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{action.Command},
	}

	resp, err := cli.ContainerCreate(ctx, config, nil, nil, action.GetName())
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}
	return resp.ID, nil
}

func getLogs(ctx context.Context, cli *client.Client, id string) (io.ReadCloser, error) {
	return cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true})
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
