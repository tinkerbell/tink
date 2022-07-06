package server

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/db/mock"
	"github.com/tinkerbell/tink/metrics"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

const (
	workflowID = "5a6d7564-d699-4e9f-a29c-a5890ccbd768"
	workerID   = "20fd5833-118f-4115-bd7b-1cf94d0f5727"
	invalidID  = "d699-4e9f-a29c-a5890ccbd"
	actionName = "install-rootfs"
	taskName   = "ubuntu-provisioning"

	defaultTestTimeout = time.Millisecond * 10
)

var wfData = []byte("{'os': 'ubuntu', 'base_url': 'http://192.168.1.1/'}")

func testServer(t *testing.T, d db.Database) *DBServer {
	t.Helper()
	l, err := log.Init("github.com/tinkerbell/tink")
	if err != nil {
		t.Errorf("log init failed: %v", err)
	}

	return &DBServer{
		logger: l,
		db:     d,
	}
}

func TestMain(m *testing.M) {
	l, _ := log.Init("github.com/tinkerbell/tink")
	metrics.SetupMetrics("onprem", l.Package("grpcserver"))
	os.Exit(m.Run())
}

func TestGetWorkflowContextList(t *testing.T) {
	type (
		args struct {
			db       *mock.DB
			workerID string
		}
		want struct {
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"empty worker id": {
			args: args{
				db: &mock.DB{},
			},
			want: want{
				expectedError: true,
			},
		},
		"database failure": {
			args: args{
				db: &mock.DB{
					GetWorkflowsForWorkerFunc: func(ctx context.Context, id string) ([]string, error) {
						return []string{workflowID}, nil
					},
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return nil, errors.New("SELECT from worflow_state")
					},
				},
				workerID: workerID,
			},
			want: want{
				expectedError: true,
			},
		},
		"no workflows found": {
			args: args{
				db: &mock.DB{
					GetWorkflowsForWorkerFunc: func(ctx context.Context, id string) ([]string, error) {
						return nil, nil
					},
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return nil, nil
					},
				},
				workerID: workerID,
			},
			want: want{
				expectedError: false,
			},
		},
		"workflows found": {
			args: args{
				db: &mock.DB{
					GetWorkflowsForWorkerFunc: func(ctx context.Context, id string) ([]string, error) {
						return []string{workflowID}, nil
					},
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
				},
				workerID: workerID,
			},
			want: want{
				expectedError: false,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			res, err := s.GetWorkflowContextList(ctx, &pb.WorkflowContextRequest{WorkerId: tc.args.workerID})
			if err != nil {
				assert.Error(t, err)
				assert.Nil(t, res)
				assert.True(t, tc.want.expectedError)
				return
			}
			if err == nil && res == nil {
				assert.False(t, tc.want.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Len(t, res.WorkflowContexts, 1)
		})
	}
}

func TestGetWorkflowActions(t *testing.T) {
	type (
		args struct {
			db         *mock.DB
			workflowID string
		}
		want struct {
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"empty workflow id": {
			args: args{
				db: &mock.DB{},
			},
			want: want{
				expectedError: true,
			},
		},
		"database failure": {
			args: args{
				db: &mock.DB{
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return nil, errors.New("SELECT from worflow_state")
					},
				},
				workflowID: invalidID,
			},
			want: want{
				expectedError: true,
			},
		},
		"getting actions": {
			args: args{
				db: &mock.DB{
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
				workflowID: workflowID,
			},
			want: want{
				expectedError: false,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			res, err := s.GetWorkflowActions(ctx, &pb.WorkflowActionsRequest{WorkflowId: tc.args.workflowID})
			if err != nil {
				assert.True(t, tc.want.expectedError)
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Len(t, res.ActionList, 1)
		})
	}
}

func TestReportActionStatus(t *testing.T) {
	type (
		args struct {
			db                                         *mock.DB
			workflowID, taskName, actionName, workerID string
			actionState                                pb.State
		}
		want struct {
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"empty workflow id": {
			args: args{
				db:         &mock.DB{},
				taskName:   taskName,
				actionName: actionName,
			},
			want: want{
				expectedError: true,
			},
		},
		"empty task name": {
			args: args{
				db:         &mock.DB{},
				workflowID: workflowID,
				actionName: actionName,
			},
			want: want{
				expectedError: true,
			},
		},
		"empty action name": {
			args: args{
				db:         &mock.DB{},
				taskName:   taskName,
				workflowID: workflowID,
			},
			want: want{
				expectedError: true,
			},
		},
		"error getting workflow context": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return nil, errors.New("SELECT from worflow_state")
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  actionName,
				actionState: pb.State_STATE_PENDING,
			},
			want: want{
				expectedError: true,
			},
		},
		"failed getting actions for context": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return nil, errors.New("SELECT from worflow_state")
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  actionName,
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: true,
			},
		},
		"success reporting status": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
					UpdateWorkflowStateFunc: func(ctx context.Context, wfContext *pb.WorkflowContext) error {
						return nil
					},
					InsertIntoWorkflowEventTableFunc: func(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
						return nil
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  actionName,
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: false,
			},
		},
		"report status for final action": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 2,
							CurrentActionState:   pb.State_STATE_RUNNING,
							CurrentAction:        actionName,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     "first-action",
									Timeout:  int64(90),
									TaskName: taskName,
								},
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
					UpdateWorkflowStateFunc: func(ctx context.Context, wfContext *pb.WorkflowContext) error {
						return nil
					},
					InsertIntoWorkflowEventTableFunc: func(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
						return nil
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  actionName,
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: false,
			},
		},
		"report status for after final action": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 2,
							CurrentActionState:   pb.State_STATE_SUCCESS,
							CurrentAction:        actionName,
							CurrentActionIndex:   2,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     "first-action",
									Timeout:  int64(90),
									TaskName: taskName,
								},
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
					UpdateWorkflowStateFunc: func(ctx context.Context, wfContext *pb.WorkflowContext) error {
						return nil
					},
					InsertIntoWorkflowEventTableFunc: func(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
						return nil
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  actionName,
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: true,
			},
		},
		"report status for second action": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionIndex:   0,
							CurrentAction:        "disk-wipe",
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    "disk-wipe",
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
					UpdateWorkflowStateFunc: func(ctx context.Context, wfContext *pb.WorkflowContext) error {
						return nil
					},
					InsertIntoWorkflowEventTableFunc: func(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
						return nil
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  actionName,
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: false,
			},
		},
		"reporting different action name": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  "different-action-name",
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: true,
			},
		},
		"reporting different task name": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    "different-task-name",
				actionName:  taskName,
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: true,
			},
		},
		"failed to update workflow state": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
					UpdateWorkflowStateFunc: func(ctx context.Context, wfContext *pb.WorkflowContext) error {
						return errors.New("INSERT in to workflow_state")
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  actionName,
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: true,
			},
		},
		"failed to update workflow events": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
					UpdateWorkflowStateFunc: func(ctx context.Context, wfContext *pb.WorkflowContext) error {
						return nil
					},
					InsertIntoWorkflowEventTableFunc: func(ctx context.Context, wfEvent *pb.WorkflowActionStatus, time time.Time) error {
						return errors.New("INSERT in to workflow_event")
					},
				},
				workflowID:  workflowID,
				workerID:    workerID,
				taskName:    taskName,
				actionName:  actionName,
				actionState: pb.State_STATE_RUNNING,
			},
			want: want{
				expectedError: true,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			res, err := s.ReportActionStatus(ctx,
				&pb.WorkflowActionStatus{
					WorkflowId:   tc.args.workflowID,
					ActionName:   tc.args.actionName,
					TaskName:     tc.args.taskName,
					WorkerId:     tc.args.workerID,
					ActionStatus: tc.args.actionState,
					Seconds:      0,
				},
			)
			if err != nil {
				assert.True(t, tc.want.expectedError)
				assert.Error(t, err)
				assert.Empty(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}

func TestUpdateWorkflowData(t *testing.T) {
	type (
		args struct {
			db         *mock.DB
			data       []byte
			workflowID string
		}
		want struct {
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"empty workflow id": {
			args: args{
				db: &mock.DB{},
			},
			want: want{
				expectedError: true,
			},
		},
		"database failure": {
			args: args{
				db: &mock.DB{
					InsertIntoWfDataTableFunc: func(ctx context.Context, req *pb.UpdateWorkflowDataRequest) error {
						return errors.New("INSERT Into workflow_data")
					},
				},
				workflowID: workflowID,
				data:       wfData,
			},
			want: want{
				expectedError: true,
			},
		},
		"add new data": {
			args: args{
				db: &mock.DB{
					InsertIntoWfDataTableFunc: func(ctx context.Context, req *pb.UpdateWorkflowDataRequest) error {
						return nil
					},
				},
				workflowID: workflowID,
				data:       wfData,
			},
			want: want{
				expectedError: false,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			res, err := s.UpdateWorkflowData(
				ctx, &pb.UpdateWorkflowDataRequest{
					WorkflowId: tc.args.workflowID,
					Data:       tc.args.data,
				})
			if err != nil {
				assert.True(t, tc.want.expectedError)
				assert.Error(t, err)
				assert.Empty(t, res)
			}
		})
	}
}

func TestGetWorkflowData(t *testing.T) {
	type (
		args struct {
			db         *mock.DB
			workflowID string
		}
		want struct {
			expectedError bool
			data          []byte
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"empty workflow id": {
			args: args{
				db: &mock.DB{
					GetfromWfDataTableFunc: func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
						return []byte{}, nil
					},
				},
				workflowID: "",
			},
			want: want{
				expectedError: true,
				data:          []byte{},
			},
		},
		"invalid workflow id": {
			args: args{
				db: &mock.DB{
					GetfromWfDataTableFunc: func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
						return []byte{}, errors.New("invalid uuid")
					},
				},
				workflowID: "d699-4e9f-a29c-a5890ccbd",
			},
			want: want{
				expectedError: true,
				data:          []byte{},
			},
		},
		"no workflow data": {
			args: args{
				db: &mock.DB{
					GetfromWfDataTableFunc: func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
						return []byte{}, nil
					},
				},
				workflowID: workflowID,
			},
			want: want{
				data: []byte{},
			},
		},
		"workflow data": {
			args: args{
				db: &mock.DB{
					GetfromWfDataTableFunc: func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
						return wfData, nil
					},
				},
				workflowID: workflowID,
			},
			want: want{
				data: wfData,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		s := testServer(t, tc.args.db)
		t.Run(name, func(t *testing.T) {
			res, err := s.GetWorkflowData(ctx, &pb.GetWorkflowDataRequest{WorkflowId: tc.args.workflowID})
			if err != nil {
				assert.True(t, tc.want.expectedError)
				assert.Error(t, err)
				assert.Equal(t, tc.want.data, res.Data)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res.Data)
			assert.Equal(t, tc.want.data, res.Data)
		})
	}
}

func TestGetWorkflowsForWorker(t *testing.T) {
	type (
		args struct {
			db       *mock.DB
			workerID string
		}
		want struct {
			data          []string
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"empty workflow id": {
			args: args{
				db:       &mock.DB{},
				workerID: "",
			},
			want: want{
				expectedError: true,
			},
		},
		"database failure": {
			args: args{
				db: &mock.DB{
					GetWorkflowsForWorkerFunc: func(ctx context.Context, id string) ([]string, error) {
						return nil, errors.New("database failed")
					},
				},
				workerID: workerID,
			},
			want: want{
				expectedError: true,
			},
		},
		"no workflows found": {
			args: args{
				db: &mock.DB{
					GetWorkflowsForWorkerFunc: func(ctx context.Context, id string) ([]string, error) {
						return nil, nil
					},
				},
				workerID: workerID,
			},
			want: want{
				expectedError: false,
			},
		},
		"workflows found": {
			args: args{
				db: &mock.DB{
					GetWorkflowsForWorkerFunc: func(ctx context.Context, id string) ([]string, error) {
						return []string{workflowID}, nil
					},
				},
				workerID: workerID,
			},
			want: want{
				data: []string{workflowID},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			res, err := getWorkflowsForWorker(context.Background(), s.db, tc.args.workerID)
			if err != nil {
				assert.True(t, tc.want.expectedError)
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want.data, res)
		})
	}
}

func TestGetWorkflowMetadata(t *testing.T) {
	type (
		args struct {
			db         *mock.DB
			workflowID string
		}
		want struct {
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"database failure": {
			args: args{
				db: &mock.DB{
					GetWorkflowMetadataFunc: func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
						return []byte{}, errors.New("SELECT from workflow_data")
					},
				},
				workflowID: workflowID,
			},
			want: want{
				expectedError: true,
			},
		},
		"no metadata": {
			args: args{
				db: &mock.DB{
					GetWorkflowMetadataFunc: func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
						return []byte{}, nil
					},
				},
				workflowID: workflowID,
			},
			want: want{
				expectedError: false,
			},
		},
		"metadata": {
			args: args{
				db: &mock.DB{
					GetWorkflowMetadataFunc: func(ctx context.Context, req *pb.GetWorkflowDataRequest) ([]byte, error) {
						type workflowMetadata struct {
							WorkerID  string    `json:"worker-id"`
							Action    string    `json:"action-name"`
							Task      string    `json:"task-name"`
							UpdatedAt time.Time `json:"updated-at"`
							SHA       string    `json:"sha256"`
						}

						meta, _ := json.Marshal(workflowMetadata{
							WorkerID:  workerID,
							Action:    actionName,
							Task:      taskName,
							UpdatedAt: time.Now(),
							SHA:       "fcbf74596047b6d3e746702ccc2c697d87817371918a5042805c8c7c75b2cb5f",
						})
						return meta, nil
					},
				},
				workflowID: workflowID,
			},
			want: want{
				expectedError: false,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			res, err := s.GetWorkflowMetadata(ctx, &pb.GetWorkflowDataRequest{WorkflowId: tc.args.workflowID})
			if err != nil {
				assert.True(t, tc.want.expectedError)
				assert.Error(t, err)
				assert.Empty(t, res.Data)
				return
			}
			if err == nil && len(res.Data) == 0 {
				assert.False(t, tc.want.expectedError)
				assert.Empty(t, res.Data)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res.Data)

			var meta map[string]string
			_ = json.Unmarshal(res.Data, &meta)
			assert.Equal(t, workerID, meta["worker-id"])
			assert.Equal(t, actionName, meta["action-name"])
			assert.Equal(t, taskName, meta["task-name"])
		})
	}
}

func TestGetWorkflowDataVersion(t *testing.T) {
	type (
		args struct {
			db *mock.DB
		}
		want struct {
			version       int32
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"database failure": {
			args: args{
				db: &mock.DB{
					GetWorkflowDataVersionFunc: func(ctx context.Context, workflowID string) (int32, error) {
						return -1, errors.New("SELECT from workflow_data")
					},
				},
			},
			want: want{
				version:       -1,
				expectedError: true,
			},
		},
		"success": {
			args: args{
				db: &mock.DB{
					GetWorkflowDataVersionFunc: func(ctx context.Context, workflowID string) (int32, error) {
						return 2, nil
					},
				},
			},
			want: want{
				version:       2,
				expectedError: false,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			res, err := s.GetWorkflowDataVersion(ctx, &pb.GetWorkflowDataRequest{WorkflowId: workflowID})
			assert.Equal(t, tc.want.version, res.Version)
			if err != nil {
				assert.True(t, tc.want.expectedError)
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestIsApplicableToSend(t *testing.T) {
	type (
		args struct {
			db *mock.DB
		}
		want struct {
			isApplicable bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"failed state": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_FAILED,
						}, nil
					},
				},
			},
			want: want{
				isApplicable: false,
			},
		},
		"timeout state": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_FAILED,
						}, nil
					},
				},
			},
			want: want{
				isApplicable: false,
			},
		},
		"failed to get actions": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_PENDING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return nil, errors.New("SELECT from worflow_state")
					},
				},
			},
			want: want{
				isApplicable: false,
			},
		},
		"is last action and success state": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_SUCCESS,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
			},
			want: want{
				isApplicable: false,
			},
		},
		"in-progress last action for different worker": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_RUNNING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: "c160ee99-a969-49d3-8415-3dbceeff54fd",
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
			},
			want: want{
				isApplicable: false,
			},
		},
		"success state and not the last action": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_SUCCESS,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    "disk-wipe",
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
			},
			want: want{
				isApplicable: true,
			},
		},
		"not the last action": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_RUNNING,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    "disk-wipe",
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
			},
			want: want{
				isApplicable: true,
			},
		},
	}

	logger, _ := log.Init("test")
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			wfContext, _ := s.db.GetWorkflowContexts(ctx, workflowID)
			res := isApplicableToSend(ctx, logger, wfContext, workerID, s.db)
			assert.Equal(t, tc.want.isApplicable, res)
		})
	}
}

func TestIsLastAction(t *testing.T) {
	type (
		args struct {
			db *mock.DB
		}
		want struct {
			isLastAction bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"is not last": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_SUCCESS,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    "disk-wipe",
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
			},
			want: want{
				isLastAction: false,
			},
		},
		"is last": {
			args: args{
				db: &mock.DB{
					GetWorkflowContextsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
						return &pb.WorkflowContext{
							WorkflowId:           workflowID,
							TotalNumberOfActions: 1,
							CurrentActionState:   pb.State_STATE_SUCCESS,
						}, nil
					},
					GetWorkflowActionsFunc: func(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
						return &pb.WorkflowActionList{
							ActionList: []*pb.WorkflowAction{
								{
									WorkerId: workerID,
									Image:    actionName,
									Name:     actionName,
									Timeout:  int64(90),
									TaskName: taskName,
								},
							},
						}, nil
					},
				},
			},
			want: want{
				isLastAction: true,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.db)
			wfContext, _ := s.db.GetWorkflowContexts(ctx, workflowID)
			actions, _ := s.db.GetWorkflowActions(ctx, workflowID)
			res := isLastAction(wfContext, actions)
			assert.Equal(t, tc.want.isLastAction, res)
		})
	}
}
