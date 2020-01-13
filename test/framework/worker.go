package framework

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dc "github.com/docker/docker/client"
	"github.com/packethost/rover/protos/workflow"
	"github.com/pkg/errors"
)

var cli *dc.Client
var workerID = []string{"f9f56dff-098a-4c5f-a51c-19ad35de85d1", "f9f56dff-098a-4c5f-a51c-19ad35de85d2"}

func initializeDockerClient() (*dc.Client, error) {
	c, err := dc.NewClientWithOpts(dc.FromEnv, dc.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "DOCKER CLIENT")
	}
	return c, nil
}

func createWorkerContainer(ctx context.Context, cli *dc.Client, workerID string) (string, error) {
	volume := map[string]struct{}{"/var/run/docker.sock": struct{}{}, "/workflow": struct{}{}}
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

func waitContainer(ctx context.Context, cli *dc.Client, id string, wg *sync.WaitGroup, failedWorkers chan<- string, statusChannel chan<- int64) {
	// send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case status := <-wait:
		statusChannel <- status.StatusCode
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
func checkCurrentStatus(ctx context.Context, wfID string, workflowStatus chan workflow.ActionState) {
	for len(workflowStatus) == 0 {
		GetCurrentStatus(ctx, wfID, workflowStatus)
	}
}

func StartWorkers(workers int64, workerStatus chan<- int64, wfID string) (workflow.ActionState, error) {

	var wg sync.WaitGroup
	failedWorkers := make(chan string, workers)
	workflowStatus := make(chan workflow.ActionState, 1)
	cli, err := initializeDockerClient()
	if err != nil {
		return workflow.ActionState_ACTION_FAILED, err
	}
	workerContainer := make([]string, workers)
	var i int64
	for i = 0; i < workers; i++ {
		ctx := context.Background()
		cID, err := createWorkerContainer(ctx, cli, workerID[i])
		if err != nil {
			fmt.Println("Worker with failed to create: ", err)
			// TODO Should be remove all the containers which previously created?
		} else {
			workerContainer[i] = cID
			fmt.Println("Worker Created with ID : ", cID)
			// Run container

			err = runContainer(ctx, cli, cID)
		}

		if err != nil {
			fmt.Println("Worker with id ", cID, " failed to start: ", err)
			// TODO Should be remove the containers which started previously
		} else {
			fmt.Println("Worker started with ID : ", cID)
			wg.Add(1)
			go waitContainer(ctx, cli, cID, &wg, failedWorkers, workerStatus)
			go checkCurrentStatus(ctx, wfID, workflowStatus)
		}
	}

	if err != nil {
		return workflow.ActionState_ACTION_FAILED, err
	}

	status := <-workflowStatus
	fmt.Println("Status of Workflow : ", status)
	wg.Wait()
	//ctx := context.Background()
	for _, cID := range workerContainer {
		//err := removeContainer(ctx, cli, cID)
		if err != nil {
			fmt.Println("Failed to remove worker container with ID : ", cID)
		}
	}

	if len(failedWorkers) > 0 {
		for i = 0; i < workers; i++ {
			failedContainer, ok := <-failedWorkers
			if ok {
				fmt.Println("Worker Failed : ", failedContainer)
				err = errors.New("Test Failed")
			}

			if len(failedContainer) > 0 {
				continue
			} else {
				break
			}
		}
	}
	if err != nil {
		return status, err
	}
	fmt.Println("Test Passed")
	return status, nil
}
