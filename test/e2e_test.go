package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	//"github.com/moby/api/types/container"

	//"github.com/moby/api/types"
	"github.com/packethost/rover/client"
	"github.com/packethost/rover/test/framework"

	//"github.com/packethost/rover/test/target"
	//"github.com/packethost/rover/test/template"
	//	"github.com/packethost/rover/test/workflow"
	"github.com/stretchr/testify/assert"
)

//var workerID []string

func TestMain(m *testing.M) {
	fmt.Println("########Creating Setup########")
	err := framework.StartStack()
	time.Sleep(10 * time.Second)
	if err != nil {
		os.Exit(1)
	}
	client.Setup()
	fmt.Println("########Setup Created########")

	fmt.Println("Creating hardware inventory")
	//push hardware data into hardware table
	hwData := []string{"hardware_1.json", "hardware_2.json"}
	err = framework.PushHardwareData(hwData)
	if err != nil {
		fmt.Println("Failed to push hardware inventory : ", err)
		os.Exit(2)
	}
	fmt.Println("Hardware inventory created")

	//framework.workerID = []string{"f9f56dff-098a-4c5f-a51c-19ad35de85d1", "f9f56dff-098a-4c5f-a51c-19ad35de85d2"}

	fmt.Println("########Starting Tests########")
	status := m.Run()
	fmt.Println("########Finished Tests########")
	fmt.Println("########Removing setup########")
	err = framework.TearDown()
	if err != nil {
		os.Exit(3)
	}
	fmt.Println("########Setup removed########")
	os.Exit(status)
}

func createWorkflow(tar string, tmpl string) error {

	//Add target machine mac/ip addr into targets table
	targetID, err := framework.CreateTargets(tar)
	if err != nil {
		return err
	}
	fmt.Println("Target Created : ", targetID)
	//Add template in template table
	templateID, err := framework.CreateTemplate(tmpl)
	if err != nil {
		return err
	}
	fmt.Println("Template Created : ", templateID)
	workflowID, err := framework.CreateWorkflow(templateID, targetID)
	if err != nil {
		return err
	}
	fmt.Println("Workflow Created : ", workflowID)
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
		assert.NoError(t, err, "Create Workflow")

		// Start the Worker
		workerStatus := make(chan int64, test.workers)
		err = framework.StartWorkers(test.workers, workerStatus)
		if err != nil {
			fmt.Printf("Test_%d with values : %v Failed\n", i, test)
			t.Error(err)
		}
		assert.NoError(t, err, "Workers Failed")

		for i := int64(0); i < test.workers; i++ {
			fmt.Println("lenght of channel is : ", len(workerStatus))
			if len(workerStatus) > 0 {
				fmt.Println("Check for worker exit status")
				status := <-workerStatus
				assert.Equal(t, int64(0), status)
			}
		}
		fmt.Printf("Test_%d with values : %v Passed\n", (i + 1), test)
	}
}
