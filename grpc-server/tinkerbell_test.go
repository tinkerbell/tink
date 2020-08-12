package grpcserver

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tinkerbell/tink/db/mock"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

const (
	invalidID            = "d699-4e9f-a29c-a5890ccbd"
	workflowWithNoData   = "5a6d7564-d699-4e9f-a29c-a5890ccbd768"
	workflowWithData     = "5711afcf-ea0b-4055-b4d6-9f88080f7afc"
	workerWithNoWorkflow = "4ebf0efa-b913-45a1-a9bf-c59829cb53a9"
	workerWithWorkflow   = "20fd5833-118f-4115-bd7b-1cf94d0f5727"
	workerForErrCases    = "b6e1a7ba-3a68-4695-9846-c5fb1eee8bee"
)

var testServer = &server{
	db: mock.DB{},
}

func TestMain(m *testing.M) {
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
			res:      []string{workflowWithNoData, workflowWithData},
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

func TestGetWorkflowData(t *testing.T) {
	testCases := []struct {
		name          string
		req           *pb.GetWorkflowDataRequest
		data          []byte
		expectedError bool
	}{
		{
			name:          "empty workflow id",
			req:           &pb.GetWorkflowDataRequest{WorkflowID: ""},
			data:          []byte{},
			expectedError: true,
		},
		{
			name:          "invalid  workflow id",
			req:           &pb.GetWorkflowDataRequest{WorkflowID: invalidID},
			data:          []byte{},
			expectedError: true,
		},
		{
			name:          "workflow id with no data",
			req:           &pb.GetWorkflowDataRequest{WorkflowID: workflowWithNoData},
			data:          []byte{},
			expectedError: false,
		},
		{
			name:          "workflow id with data",
			req:           &pb.GetWorkflowDataRequest{WorkflowID: workflowWithData},
			data:          []byte("{'os': 'ubuntu', 'base_url': 'http://192.168.1.1/'}"),
			expectedError: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := testServer.GetWorkflowData(context.TODO(), test.req)

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
