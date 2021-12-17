package tink

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
)

type mockCli struct {
	pb.WorkflowServiceClient
	response     pb.WorkflowService_GetWorkflowContextsClient
	actionstatus pb.WorkflowActionStatus
}

func (mcli *mockCli) GetWorkflowContexts(_ context.Context, _ *pb.WorkflowContextRequest, _ ...grpc.CallOption) (pb.WorkflowService_GetWorkflowContextsClient, error) {
	return mcli.response, nil
}

func (mcli *mockCli) GetWorkflowActions(_ context.Context, _ *pb.WorkflowActionsRequest, _ ...grpc.CallOption) (*pb.WorkflowActionList, error) {
	return nil, nil
}

func (mcli *mockCli) ReportActionStatus(_ context.Context, _ *pb.WorkflowActionStatus, _ ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}

func TestWorkerGetWorkflowActions(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx context.Context
		cli pb.WorkflowServiceClient
	}
	tests := []struct {
		name    string
		data    args
		w       Wdata
		wantErr error
	}{
		{
			name: "get_workflow_actions_with_workflowID",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					WorkflowID: "3431423",
				},
			},
			wantErr: nil,
		},
		{
			name: "get_workflow_actions_with_no_workflowID",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					WorkflowID: "",
				},
			},
			wantErr: errors.New("Empty string is not a valid workflow id"),
		},
		{
			name: "get_workflow_actions_with_no_service_client",
			data: args{
				ctx: context.Background(),
				cli: nil,
			},
			w: Wdata{
				Data: WorkflowData{
					WorkflowID: "3432315",
				},
			},
			wantErr: errors.New("WorkflowServiceClient interface is not valid"),
		},
		{
			name: "get_workflow_actions_with_service_client",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					WorkflowID: "3432315",
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.w.GetWorkflowActions(tt.data.ctx, tt.data.cli)
			if err != nil {
				diff := cmp.Diff(tt.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestWorkerReportActionStatus(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx context.Context
		cli pb.WorkflowServiceClient
	}
	tests := []struct {
		name    string
		data    args
		w       Wdata
		wantErr error
	}{
		{
			name: "get_workflow_actions_with_no_service_client",
			data: args{
				ctx: context.Background(),
				cli: nil,
			},
			w: Wdata{
				Data: WorkflowData{
					ActionStatus: &mock.actionstatus,
				},
			},
			wantErr: errors.New("WorkflowServiceClient interface is not valid"),
		},
		{
			name: "get_workflow_actions_with_service_client",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					ActionStatus: &mock.actionstatus,
				},
			},
			wantErr: nil,
		},
		{
			name: "get_workflow_actions_with_no_action_status",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					ActionStatus: nil,
				},
			},
			wantErr: errors.New("WorkflowActionStatus is not valid"),
		},
		{
			name: "get_workflow_actions_with_action_status",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					ActionStatus: &mock.actionstatus,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.w.ReportActionStatus(tt.data.ctx, tt.data.cli)
			if err != nil {
				diff := cmp.Diff(tt.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestWorkerGetWorkflowContexts(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx context.Context
		cli pb.WorkflowServiceClient
	}
	tests := []struct {
		name    string
		data    args
		w       Wdata
		wantErr error
	}{
		{
			name: "get_workflow_actions_with_no_service_client",
			data: args{
				ctx: context.Background(),
				cli: nil,
			},
			w: Wdata{
				Data: WorkflowData{
					WorkerID: "3432315",
				},
			},
			wantErr: errors.New("WorkflowServiceClient interface is not valid"),
		},
		{
			name: "get_workflow_actions_with_service_client",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					WorkerID: "3432315",
				},
			},
			wantErr: nil,
		},
		{
			name: "get_workflow_actions_with_workerID",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					WorkerID: "3432315",
				},
			},
			wantErr: nil,
		},
		{
			name: "get_workflow_actions_with_no_workerID",
			data: args{
				ctx: context.Background(),
				cli: mock,
			},
			w: Wdata{
				Data: WorkflowData{
					WorkerID: "",
				},
			},
			wantErr: errors.New("Empty string is not a valid worker id"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.w.GetWorkflowContexts(tt.data.ctx, tt.data.cli)
			if err != nil {
				diff := cmp.Diff(tt.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
