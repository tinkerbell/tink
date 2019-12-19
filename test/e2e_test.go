package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	//"github.com/moby/api/types/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dc "github.com/docker/docker/client"

	//"github.com/moby/api/types"
	"github.com/packethost/rover/client"
	"github.com/packethost/rover/test/hardware"
	"github.com/packethost/rover/test/target"
	"github.com/packethost/rover/test/template"
	"github.com/packethost/rover/test/workflow"
	"github.com/pkg/errors"
)

func startDb(filepath string) error {
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d db")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func startStack() error {
	// Docker compose file for starting the containers
	filepath := os.Getenv("GOPATH") + "src/github.com/packethost/rover/docker-compose.yml "

	// Start Db and logging components
	err := startDb(filepath)
	if err != nil {
		return err
	}

	// Wait for some time so thath the above containers to be in running condition
	time.Sleep(5 * time.Second)

	// Start other containers
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func TestMain(m *testing.M) {
	fmt.Println("Inside Main")
	err := startStack()
	time.Sleep(3 * time.Second)
	client.Setup()
	if err != nil {
		os.Exit(1)
	}
	status := m.Run()
	fmt.Println("removing setup")
	//err = tearDown()
	if err != nil {
		os.Exit(2)
	}
	os.Exit(status)
}

func createWorkflow() error {
	//push hardware data into hardware table
	err := hardware.PushHardwareData()
	if err != nil {
		return err
	}
	fmt.Println("Hardware Data pushed for ID : f9f56dff-098a-4c5f-a51c-19ad35de85d1")
	//Add target machine mac/ip addr into targets table
	targetID, err := target.CreateTargets()
	if err != nil {
		return err
	}
	fmt.Println("Target Created : ", targetID)
	//Add template in template table
	templateID, err := template.CreateTemplate()
	if err != nil {
		return err
	}
	fmt.Println("Template Created : ", templateID)
	workflowID, err := workflow.CreateWorkflow(templateID, targetID)
	if err != nil {
		return err
	}
	fmt.Println("Workflow Created : ", workflowID)
	return nil
}

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
		Env:          []string{"ROVER_GRPC_AUTHORITY=127.0.0.1:42113", "ROVER_CERT_URL=http://127.0.0.1:42114/cert", "WORKER_ID=f9f56dff-098a-4c5f-a51c-19ad35de85d1", "DOCKER_REGISTRY=127.0.0.1:5000", "DOCKER_API_VERSION=v1.40"},
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

func waitContainer(ctx context.Context, cli *dc.Client, id string) (int64, error) {
	// send API call to wait for the container completion
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case status := <-wait:
		return status.StatusCode, nil
	case err := <-errC:
		return 1, err
	}
}

func startWorkers(workers int64) error {
	cli, err := initializeDockerClient()
	if err != nil {
		return err
	}

	workerID := []string{"f9f56dff-098a-4c5f-a51c-19ad35de85d1"}
	var i int64
	for i = 0; i < workers; i++ {
		ctx := context.Background()
		cID, err := createWorkerContainer(ctx, cli, workerID[i])
		fmt.Println("Container Created with ID : ", cID)
		// Run container
		err = runContainer(ctx, cli, cID)
		if err != nil {
			return err
		}
	}
	//err = waitContainer(ctx, )
	return nil
}
func TestRover(t *testing.T) {

	// Create Workflow for first case
	err := createWorkflow()
	if err != nil {
		t.Error(err)
	}

	// Start the Worker
	err = startWorkers(1)
	if err != nil {
		t.Error(err)
	}
}
