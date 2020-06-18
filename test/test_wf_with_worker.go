package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
	"github.com/tinkerbell/tink/test/framework"
)

// TestWfWithWorker : One Worker Test
func TestWfWithWorker(t *testing.T) {

	// Start test only if the test case exist in the table
	if test, ok := testCases["testWfWithWorker"]; ok {
		wfID, err := framework.SetupWorkflow(test.hMAC, test.template)

		if err != nil {
			t.Error(err)
		}
		if !assert.NoError(t, err, "Create Workflow") {
			t.Fatal(err)
		}

		// Start the Worker
		workerStatus := make(chan int64, test.workers)
		wfStatus, err := framework.StartWorkers(test.workers, workerStatus, wfID)
		if err != nil {
			log.Info("Test Failed")
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
				if err == nil {
					log.Info("Comparing EPH data")
					assert.Equal(t, test.ephData, string(resp.GetData()))
				} else {
					log.Error("Get EPH data failed")
					log.Error(err)
				}
			}
		}
	}
}
