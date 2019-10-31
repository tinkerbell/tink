package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	pb "github.com/packethost/rover/protos/rover"
)

var (
	registry string
	cli      *client.Client
)

func executeAction(ctx context.Context, action *pb.WorkflowAction) error {
	id, err := createContainer(ctx, action)
	if err != nil {
		log.Fatalln("Action: ", action.Name, "Failed for error: ", err)
		return err
	}

	out, err := getLogs(ctx, cli, id)
	defer out.Close()
	if err != nil {
		log.Fatalln("Action: ", action.Name, "Failed for error: ", err)
		return err
	}
	// TODO: use "github.com/docker/docker/pkg/stdcopy"
	io.Copy(os.Stdout, out)
	return nil
}

func createContainer(ctx context.Context, action *pb.WorkflowAction) (string, error) {
	config := &container.Config{
		Image:        registry + "/" + action.GetImage(),
		AttachStdout: true,
		AttachStderr: true,
	}
	resp, err := cli.ContainerCreate(ctx, config, nil, nil, action.GetName())
	if err != nil {
		log.Fatalln(err)
		return "", err
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	return resp.ID, nil
}

func getLogs(ctx context.Context, cli *client.Client, id string) (io.ReadCloser, error) {
	return cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true})
}

func init() {
	registry = os.Getenv("DOCKER_REGISTRY")
	if registry == "" {
		log.Fatalln(fmt.Errorf("requried DOCKER_REGISTRY"))
	}
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion(os.Getenv("DOCKER_API_VERSION")))
	if err != nil {
		log.Fatalln(err)
	}
	cli = c
}
