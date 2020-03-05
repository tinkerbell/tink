package e2e

import (
	"context"
	"testing"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/workflow"
	"github.com/packethost/rover/test/framework"
	"github.com/stretchr/testify/assert"
)

// TestWfWithWorker : One Worker Test
func TestWfWithWorker(t *testing.T) {

	// Start test only if the test case exist in the table
	if test, ok := testCases["testWfWithWorker"]; ok {
		wfID, err := framework.SetupWorkflow(test.target, test.template)

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
			log.Errorf("Test Failed\n")
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
