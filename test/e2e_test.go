package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/workflow"
	"github.com/packethost/rover/test/framework"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	fmt.Println("########Creating Setup########")
	err := framework.StartStack()
	time.Sleep(10 * time.Second)
	if err != nil {
		os.Exit(1)
	}
	os.Setenv("ROVER_GRPC_AUTHORITY", "127.0.0.1:42113")
	os.Setenv("ROVER_CERT_URL", "http://127.0.0.1:42114/cert")
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

	fmt.Println("########Starting Tests########")
	status := m.Run()
	fmt.Println("########Finished Tests########")
	fmt.Println("########Removing setup########")
	//err = framework.TearDown()
	if err != nil {
		os.Exit(3)
	}
	fmt.Println("########Setup removed########")
	os.Exit(status)
}

var testCases = []struct {
	//name     string
	target   string
	template string
	workers  int64
	expected workflow.ActionState
	ephData  string
}{
	{"target_1.json", "sample_1", 1, workflow.ActionState_ACTION_SUCCESS, `{"action_02": "data_02}`},
	{"target_1.json", "sample_2", 1, workflow.ActionState_ACTION_TIMEOUT, `{"action_01": "data_01}`},
}

func TestOneWorker(t *testing.T) {

	// Start test
	if len(testCases) > 0 {
		test := testCases[0]
		wfID, err := framework.SetupWorkflow(test.target, test.template)

		if err != nil {
			t.Error(err)
		}
		assert.NoError(t, err, "Create Workflow")

		// Start the Worker
		workerStatus := make(chan int64, test.workers)
		wfStatus, err := framework.StartWorkers(test.workers, workerStatus, wfID)
		if err != nil {
			fmt.Printf("Test Failed\n")
			t.Error(err)
		}
		assert.Equal(t, test.expected, wfStatus)
		assert.NoError(t, err, "Workers Failed")

		for i := int64(0); i < test.workers; i++ {
			if len(workerStatus) > 0 {
				//Check for worker exit status
				status := <-workerStatus
				expected := 0
				if test.expected != workflow.ActionState_ACTION_SUCCESS {
					expected = 1
				}
				assert.Equal(t, int64(expected), status)
				//checking for ephemeral data validation
				resp, err := client.WorkflowClient.GetWorkflowData(context.Background(), &workflow.GetWorkflowDataRequest{WorkflowID: wfID, Version: 0})
				if err != nil {
					assert.Equal(t, test.ephData, string(resp.GetData()))
				}
			}
		}
	}
}

func TestTimeout(t *testing.T) {
	// Start test
	if len(testCases) > 1 {
		test := testCases[1]
		wfID, err := framework.SetupWorkflow(test.target, test.template)

		if err != nil {
			t.Error(err)
		}
		assert.NoError(t, err, "Create Workflow")

		// Start the Worker
		workerStatus := make(chan int64, test.workers)
		wfStatus, err := framework.StartWorkers(test.workers, workerStatus, wfID)
		if err != nil {
			fmt.Printf("Test Failed\n")
			t.Error(err)
		}
		assert.Equal(t, test.expected, wfStatus)
		assert.NoError(t, err, "Workers Failed")

		for i := int64(0); i < test.workers; i++ {
			if len(workerStatus) > 0 {
				// Check for worker exit status
				status := <-workerStatus
				expected := 0
				if test.expected != workflow.ActionState_ACTION_SUCCESS {
					expected = 1
				}
				assert.Equal(t, int64(expected), status)

				//checking for ephemeral data validation
				resp, err := client.WorkflowClient.GetWorkflowData(context.Background(), &workflow.GetWorkflowDataRequest{WorkflowID: wfID, Version: 0})
				if err != nil {
					assert.Equal(t, test.ephData, string(resp.GetData()))
				}
			}
		}
	}
}
