package e2e

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dc "github.com/docker/docker/client"
	"github.com/pkg/errors"
)

func initializeDockerClient() (*dc.Client, error) {
	c, err := dc.NewClientWithOpts(dc.FromEnv, dc.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "DOCKER CLIENT")
	}
	return c, nil
}

func createWorkerContainer(ctx context.Context, cli *dc.Client, workerID string) (string, error) {
	volume := map[string]struct{}{"/var/run/docker.sock": struct{}{}}
	config := &container.Config{
		Image:        "worker",
		AttachStdout: true,
		AttachStderr: true,
		Volumes:      volume,
		Env:          []string{"ROVER_GRPC_AUTHORITY=127.0.0.1:42113", "ROVER_CERT_URL=http://127.0.0.1:42114/cert", "WORKER_ID=" + workerID, "DOCKER_REGISTRY=127.0.0.1:5000", "DOCKER_API_VERSION=v1.40"},
	}
	hostConfig := &container.HostConfig{
		NetworkMode: "host",
		Binds:       []string{"/var/run/docker.sock:/var/run/docker.sock:rw"},
	}
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, workerID)
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}
	return resp.ID, nil
}

func runContainer(ctx context.Context, cli *dc.Client, id string) error {
	err := cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "DOCKER START")
	}
	return nil
}

func waitContainer(ctx context.Context, cli *dc.Client, id string, wg *sync.WaitGroup, failedWorkers chan<- string) {
	// send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case status := <-wait:
		fmt.Println("Worker with id ", id, "finished sucessfully with status code ", status.StatusCode)
	case err := <-errC:
		fmt.Println("Worker with id ", id, "failed : ", err)
		failedWorkers <- id
	}
	wg.Done()
}

func removeContainer(ctx context.Context, cli *dc.Client, id string) error {
	// create options for removing container
	opts := types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}
	// send API call to remove the container
	err := cli.ContainerRemove(ctx, id, opts)
	if err != nil {
		return err
	}
	fmt.Println("Worker Container removed : ", id)
	return nil
}
