package internal

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/packethost/pkg/log"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

func TestNewWorker(t *testing.T) {
	type args struct {
		client        pb.WorkflowServiceClient
		regConn       *RegistryConnDetails
		logger        log.Logger
		registry      string
		retries       int
		retryInterval time.Duration
		maxFileSize   int64
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 *Worker
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := NewWorker(tArgs.client, tArgs.regConn, tArgs.logger, tArgs.registry, tArgs.retries, tArgs.retryInterval, tArgs.maxFileSize)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewWorker got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestWorker_captureLogs(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *Worker
		inspect func(r *Worker, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			receiver.captureLogs(tArgs.ctx, tArgs.id)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

		})
	}
}

func TestWorker_execute(t *testing.T) {
	type args struct {
		ctx         context.Context
		wfID        string
		action      *pb.WorkflowAction
		captureLogs bool
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *Worker
		inspect func(r *Worker, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		want1      pb.State
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			got1, err := receiver.execute(tArgs.ctx, tArgs.wfID, tArgs.action, tArgs.captureLogs)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Worker.execute got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Worker.execute error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func TestWorker_ProcessWorkflowActions(t *testing.T) {
	type args struct {
		ctx               context.Context
		workerID          string
		captureActionLogs bool
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *Worker
		inspect func(r *Worker, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			err := receiver.ProcessWorkflowActions(tArgs.ctx, tArgs.workerID, tArgs.captureActionLogs)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Worker.ProcessWorkflowActions error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_exitWithGrpcError(t *testing.T) {
	type args struct {
		err error
		l   log.Logger
	}
	tests := []struct {
		name string
		args func(t *testing.T) args
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			exitWithGrpcError(tArgs.err, tArgs.l)

		})
	}
}

func Test_isLastAction(t *testing.T) {
	type args struct {
		wfContext *pb.WorkflowContext
		actions   *pb.WorkflowActionList
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 bool
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := isLastAction(tArgs.wfContext, tArgs.actions)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("isLastAction got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestWorker_reportActionStatus(t *testing.T) {
	type args struct {
		ctx          context.Context
		actionStatus *pb.WorkflowActionStatus
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *Worker
		inspect func(r *Worker, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			err := receiver.reportActionStatus(tArgs.ctx, tArgs.actionStatus)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Worker.reportActionStatus error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_getWorkflowData(t *testing.T) {
	type args struct {
		ctx        context.Context
		logger     log.Logger
		client     pb.WorkflowServiceClient
		workerID   string
		workflowID string
	}
	tests := []struct {
		name string
		args func(t *testing.T) args
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			getWorkflowData(tArgs.ctx, tArgs.logger, tArgs.client, tArgs.workerID, tArgs.workflowID)

		})
	}
}

func TestWorker_updateWorkflowData(t *testing.T) {
	type args struct {
		ctx          context.Context
		actionStatus *pb.WorkflowActionStatus
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *Worker
		inspect func(r *Worker, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			receiver.updateWorkflowData(tArgs.ctx, tArgs.actionStatus)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

		})
	}
}

func Test_sendUpdate(t *testing.T) {
	type args struct {
		ctx      context.Context
		logger   log.Logger
		client   pb.WorkflowServiceClient
		st       *pb.WorkflowActionStatus
		data     []byte
		checksum string
	}
	tests := []struct {
		name string
		args func(t *testing.T) args
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			sendUpdate(tArgs.ctx, tArgs.logger, tArgs.client, tArgs.st, tArgs.data, tArgs.checksum)

		})
	}
}

func Test_openDataFile(t *testing.T) {
	type args struct {
		wfDir string
		l     log.Logger
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 *os.File
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := openDataFile(tArgs.wfDir, tArgs.l)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("openDataFile got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func Test_isValidDataFile(t *testing.T) {
	type args struct {
		f       *os.File
		maxSize int64
		data    []byte
		l       log.Logger
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 bool
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := isValidDataFile(tArgs.f, tArgs.maxSize, tArgs.data, tArgs.l)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("isValidDataFile got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}
