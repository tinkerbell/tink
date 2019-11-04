package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
	err := pullActionImage(ctx, action.GetImage())
	if err != nil {
		log.Fatalln("Action: ", action.Name, "Failed for error: ", err)
		return err
	}
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

	io.Copy(os.Stdout, out)
	return nil
}

func pullActionImage(ctx context.Context, image string) error {
	f, err := os.Open("/" + registry + "/ca.crt")
	defer f.Close()
	if err != nil {
		log.Fatalln("Failed to LOAD the certificate for registry:", registry)
		return err
	}
	auth, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln("Failed to READ the certificate for registry:", registry)
		return err
	}

	rw, err := cli.ImagePull(ctx, registry+"/"+image, types.ImagePullOptions{RegistryAuth: string(auth)})
	defer rw.Close()
	if err != nil {
		log.Fatalln("Failed to pull Docker image", err)
		return err
	}
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
