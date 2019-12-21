package e2e

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	//"github.com/moby/api/types/container"

	//"github.com/moby/api/types"
	"github.com/packethost/rover/client"
	"github.com/packethost/rover/test/hardware"
	"github.com/packethost/rover/test/target"
	"github.com/packethost/rover/test/template"
	"github.com/packethost/rover/test/workflow"
	"github.com/pkg/errors"
)

var workerID []string

func TestMain(m *testing.M) {
	fmt.Println("########Creating Setup########")
	err := startStack()
	time.Sleep(10 * time.Second)
	if err != nil {
		os.Exit(1)
	}
	client.Setup()
	fmt.Println("########Setup Created########")

	fmt.Println("Creating hardware inventory")
	//push hardware data into hardware table
	hwData := []string{"hardware_1.json", "hardware_2.json"}
	err = hardware.PushHardwareData(hwData)
	if err != nil {
		os.Exit(2)
	}
	fmt.Println("Hardware inventory created")

	workerID = []string{"f9f56dff-098a-4c5f-a51c-19ad35de85d1", "f9f56dff-098a-4c5f-a51c-19ad35de85d2"}

	fmt.Println("########Starting Tests########")
	status := m.Run()
	fmt.Println("########Finished Tests########")
	fmt.Println("########Removing setup########")
	err = tearDown()
	if err != nil {
		os.Exit(3)
	}
	fmt.Println("########Setup removed########")
	os.Exit(status)
}

func createWorkflow(tar string, tmpl string) error {

	//Add target machine mac/ip addr into targets table
	targetID, err := target.CreateTargets(tar)
	if err != nil {
		return err
	}
	fmt.Println("Target Created : ", targetID)
	//Add template in template table
	templateID, err := template.CreateTemplate(tmpl)
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

func startWorkers(workers int64) error {

	var wg sync.WaitGroup
	failedWorkers := make(chan string, workers)
	cli, err := initializeDockerClient()
	if err != nil {
		return err
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
			go waitContainer(ctx, cli, cID, &wg, failedWorkers)
			//go removeContainer(ctx, cli, cID, &wg)
		}
	}
	if err != nil {
		return err
	}

	wg.Wait()
	ctx := context.Background()
	for _, cID := range workerContainer {
		err := removeContainer(ctx, cli, cID)
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
		return err
	}
	fmt.Println("Test Passed")
	return nil
}

var testCases = []struct {
	target   string
	template string
	workers  int64
}{
	{"target_1.json", "sample_1", 1},
	{"target_1.json", "sample_2", 2},
}

func TestRover(t *testing.T) {

	// Start test
	for i, test := range testCases {
		fmt.Printf("Starting Test_%d with values : %v\n", (i + 1), test)
		err := createWorkflow(test.target, test.template)

		if err != nil {
			t.Error(err)
		}
		// Start the Worker
		err = startWorkers(test.workers)
		if err != nil {
			fmt.Printf("Test_%d with values : %v Failed\n", i, test)
			t.Error(err)
		}
		fmt.Printf("Test_%d with values : %v Passed\n", (i + 1), test)
	}
}
