package grpcserver

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/packethost/pkg/log"
	"github.com/stretchr/testify/assert"

	"github.com/tinkerbell/tink/db/mock"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

const (
	invalidID            = "d699-4e9f-a29c-a5890ccbd"
	workflowForErr       = "1effe50d-3f21-4083-afa4-0e1620087d99"
	firstWorkflowID      = "5a6d7564-d699-4e9f-a29c-a5890ccbd768"
	secondWorkflowID     = "5711afcf-ea0b-4055-b4d6-9f88080f7afc"
	workerWithNoWorkflow = "4ebf0efa-b913-45a1-a9bf-c59829cb53a9"
	workerWithWorkflow   = "20fd5833-118f-4115-bd7b-1cf94d0f5727"
	workerForErrCases    = "b6e1a7ba-3a68-4695-9846-c5fb1eee8bee"
	firstActionName      = "disk-wipe"
	secondActionName     = "install-rootfs"
	taskName             = "ubuntu-provisioning"
)

var (
	testServer = &server{
		db: mock.DB{},
	}
	wfData = []byte("{'os': 'ubuntu', 'base_url': 'http://192.168.1.1/'}")
)

func TestMain(m *testing.M) {
	os.Setenv("TINKERBELL_ENV", "test")
	os.Setenv("TINKERBELL_VERSION", "ignored")
	os.Setenv("ROLLBAR_TOKEN", "ignored")

	l, _, _ := log.Init("github.com/tinkerbell/tink")
	logger = l.Package("grpcserver")

	os.Exit(m.Run())
}

func TestGetWorkflowContextList(t *testing.T) {
	testCases := []struct {
		name          string
		workerID      string
		expectedError bool
	}{
		{
			name:          "empty workflow id",
			expectedError: true,
		},
		{
			name:          "database failure",
			expectedError: true,
			workerID:      workerForErrCases,
		},
		{
			name:     "no workflows found",
			workerID: workerWithNoWorkflow,
		},
		{
			name:     "workflows found for worker",
			workerID: workerWithWorkflow,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := testServer.GetWorkflowContextList(
				context.TODO(), &pb.WorkflowContextRequest{WorkerId: test.workerID},
			)
			if err != nil && test.expectedError {
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			if test.workerID == workerWithNoWorkflow {
				assert.Nil(t, res)
				return
			}
			assert.NotNil(t, res)
			assert.Len(t, res.WorkflowContexts, 2)
		})
	}
}

func TestGetWorkflowActions(t *testing.T) {
	testCases := []struct {
		name          string
		workflowID    string
		expectedError bool
	}{
		{
			name:          "empty workflow id",
			expectedError: true,
		},
		{
			name:          "invalid  workflow id",
			workflowID:    invalidID,
			expectedError: true,
		},
		{
			name:       "getting actions",
			workflowID: secondWorkflowID,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := testServer.GetWorkflowActions(
				context.TODO(), &pb.WorkflowActionsRequest{WorkflowId: test.workflowID},
			)
			if err != nil && test.expectedError {
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Len(t, res.ActionList, 1)
			assert.Len(t, res.ActionList[0].Volumes, 3)
			assert.Equal(t, res.ActionList[0].Name, secondActionName)
		})
	}
}

func TestReportActionStatus(t *testing.T) {
	type req struct {
		workflowID, taskName, actionName, workerID string
		actionState                                pb.ActionState
	}
	testCases := []struct {
		req
		name          string
		expectedError bool
	}{
		{
			name:          "empty workflow id",
			expectedError: true,
			req: req{
				taskName:   taskName,
				actionName: firstActionName,
			},
		},
		{
			name:          "empty task name",
			expectedError: true,
			req: req{
				workflowID: firstWorkflowID,
				actionName: firstActionName,
			},
		},
		{
			name:          "empty action name",
			expectedError: true,
			req: req{
				workflowID: firstWorkflowID,
				taskName:   taskName,
			},
		},
		{
			name:          "error fetching workflow context",
			expectedError: true,
			req: req{
				workflowID:  invalidID,
				workerID:    workerWithWorkflow,
				taskName:    taskName,
				actionName:  firstActionName,
				actionState: pb.ActionState_ACTION_PENDING,
			},
		},
		{
			name: "fetch workflow context",
			req: req{
				workflowID:  firstWorkflowID,
				workerID:    workerWithWorkflow,
				taskName:    taskName,
				actionName:  secondActionName,
				actionState: pb.ActionState_ACTION_IN_PROGRESS,
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := testServer.ReportActionStatus(context.TODO(),
				&pb.WorkflowActionStatus{
					WorkflowId:   test.req.workflowID,
					ActionName:   test.req.actionName,
					TaskName:     test.req.taskName,
					WorkerId:     test.req.workerID,
					ActionStatus: test.req.actionState,
					Seconds:      0,
				},
			)

			if err != nil && test.expectedError {
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}

func TestUpdateWorkflowData(t *testing.T) {
	testCases := []struct {
		name, workflowID string
		data, metadata   []byte
		expectedError    bool
	}{
		{
			name:          "empty workflow id",
			expectedError: true,
		},
		{
			name:       "add new workflow data",
			workflowID: firstWorkflowID,
			data:       wfData,
		},
		{
			name:       "update workflow data",
			workflowID: secondWorkflowID,
			data:       wfData,
		},
		{
			name:       "database failure",
			workflowID: workflowForErr,
			data:       wfData,
		},
	}
	workflowData[secondWorkflowID] = 1
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := testServer.UpdateWorkflowData(
				context.TODO(), &pb.UpdateWorkflowDataRequest{
					WorkflowID: test.workflowID,
					Data:       test.data,
					Metadata:   test.metadata,
				})
			if err != nil && test.expectedError {
				assert.Error(t, err)
				assert.Empty(t, res)
			}
		})
	}
}

func TestGetWorkflowData(t *testing.T) {
	testCases := []struct {
		name, workflowID string
		data             []byte
		expectedError    bool
	}{
		{
			name:          "empty workflow id",
			data:          []byte{},
			expectedError: true,
		},
		{
			name:          "invalid  workflow id",
			workflowID:    invalidID,
			data:          []byte{},
			expectedError: true,
		},
		{
			name:          "workflow id with no data",
			workflowID:    secondWorkflowID,
			data:          []byte{},
			expectedError: false,
		},
		{
			name:          "workflow id with data",
			workflowID:    firstWorkflowID,
			data:          wfData,
			expectedError: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := testServer.GetWorkflowData(
				context.TODO(), &pb.GetWorkflowDataRequest{WorkflowID: test.workflowID},
			)
			if err != nil && test.expectedError {
				assert.Error(t, err)
				assert.Equal(t, test.data, res.Data)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res.Data)
			assert.Equal(t, test.data, res.Data)
		})
	}
}

func TestGetWorkflowsForWorker(t *testing.T) {
	testCases := []struct {
		name          string
		workerID      string
		res           []string
		expectedError bool
	}{
		{
			name:          "empty workflow id",
			expectedError: true,
		},
		{
			name:     "no workflows found",
			workerID: workerWithNoWorkflow,
		},
		{
			name:     "workflows found for worker",
			workerID: workerWithWorkflow,
			res:      []string{firstWorkflowID, secondWorkflowID},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := getWorkflowsForWorker(testServer.db, test.workerID)
			if err != nil && test.expectedError {
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.res, res)
		})
	}
}

func TestGetWorkflowMetadata(t *testing.T) {
	testCases := []struct {
		name, workflowID string
		data             []byte
		expectedError    bool
	}{
		{
			name:          "database failure",
			workflowID:    workflowForErr,
			data:          []byte{},
			expectedError: true,
		},
		{
			name:       "workflow with no metadata",
			workflowID: firstWorkflowID,
			data:       []byte{},
		},
		{
			name:       "workflow with metadata",
			workflowID: secondWorkflowID,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := testServer.GetWorkflowMetadata(
				context.TODO(), &pb.GetWorkflowDataRequest{WorkflowID: test.workflowID},
			)
			if err != nil && test.expectedError {
				assert.Error(t, err)
				assert.Equal(t, test.data, res.Data)
				return
			}
			if err == nil && test.workflowID == firstWorkflowID {
				assert.NoError(t, err)
				assert.Equal(t, test.data, res.Data)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res.Data)

			var meta map[string]string
			_ = json.Unmarshal(res.Data, &meta)
			assert.Equal(t, workerWithWorkflow, meta["worker-id"])
			assert.Equal(t, secondActionName, meta["action-name"])
			assert.Equal(t, taskName, meta["task-name"])
		})
	}
}

func TestGetWorkflowDataVersion(t *testing.T) {
	testCases := []struct {
		name, workflowID string
		version          int32
		expectedError    bool
	}{
		{
			name:          "database failure",
			workflowID:    workflowForErr,
			version:       -1,
			expectedError: true,
		},
		{
			name:       "workflow with no data",
			workflowID: secondWorkflowID,
		},
		{
			name:       "workflow with data",
			version:    2,
			workflowID: firstWorkflowID,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := testServer.GetWorkflowDataVersion(
				context.TODO(), &pb.GetWorkflowDataRequest{WorkflowID: test.workflowID},
			)
			if err != nil && test.expectedError {
				assert.Error(t, err)
				assert.Equal(t, test.version, res.Version)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.version, res.Version)
		})
	}
}

func TestIsApplicableToSend(t *testing.T) {
	testCases := []struct {
		name, workerID, workflowID string
		actionState                pb.ActionState
		isApplicable               bool
	}{
		{
			name:        "workflow in failed state",
			workflowID:  firstWorkflowID,
			actionState: pb.ActionState_ACTION_FAILED,
		},
		{
			name:        "workflow in timeout state",
			workflowID:  firstWorkflowID,
			actionState: pb.ActionState_ACTION_TIMEOUT,
		},
		{
			name:        "is last action with success state",
			workflowID:  secondWorkflowID,
			actionState: pb.ActionState_ACTION_SUCCESS,
		},
		{
			name:         "with success state but not the last action",
			workflowID:   firstWorkflowID,
			actionState:  pb.ActionState_ACTION_SUCCESS,
			workerID:     workerWithWorkflow,
			isApplicable: true,
		},
		{
			name:         "not the last action",
			workflowID:   firstWorkflowID,
			workerID:     workerWithWorkflow,
			isApplicable: true,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			wfContext, _ := testServer.db.GetWorkflowContexts(context.TODO(), test.workflowID)
			wfContext.CurrentActionState = test.actionState
			res := isApplicableToSend(
				context.TODO(), wfContext, test.workerID, testServer.db,
			)
			assert.Equal(t, test.isApplicable, res)
		})
	}
}

func TestIsLastAction(t *testing.T) {
	testCases := []struct {
		name, workflowID string
		isLastAction     bool
	}{
		{
			name:         "not the last action",
			workflowID:   firstWorkflowID,
			isLastAction: false,
		},
		{
			name:         "is the last action",
			workflowID:   secondWorkflowID,
			isLastAction: true,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			wfContext, _ := testServer.db.GetWorkflowContexts(context.TODO(), test.workflowID)
			actions, _ := testServer.db.GetWorkflowActions(context.TODO(), test.workflowID)
			res := isLastAction(wfContext, actions)
			assert.Equal(t, test.isLastAction, res)
		})
	}
}
