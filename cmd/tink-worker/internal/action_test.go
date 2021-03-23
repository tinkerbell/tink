package internal

import (
	"context"
	"reflect"
	"testing"

	"github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

func TestWorker_createContainer(t *testing.T) {
	type args struct {
		ctx         context.Context
		cmd         []string
		wfID        string
		action      *pb.WorkflowAction
		captureLogs bool
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *Worker
		inspect func(r *Worker, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		want1      string
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			got1, err := receiver.createContainer(tArgs.ctx, tArgs.cmd, tArgs.wfID, tArgs.action, tArgs.captureLogs)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Worker.createContainer got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Worker.createContainer error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_startContainer(t *testing.T) {
	type args struct {
		ctx context.Context
		l   log.Logger
		cli *client.Client
		id  string
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			err := startContainer(tArgs.ctx, tArgs.l, tArgs.cli, tArgs.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("startContainer error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_waitContainer(t *testing.T) {
	type args struct {
		ctx context.Context
		cli *client.Client
		id  string
	}
	tests := []struct {
		name string
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

			got1, err := waitContainer(tArgs.ctx, tArgs.cli, tArgs.id)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("waitContainer got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("waitContainer error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_waitFailedContainer(t *testing.T) {
	type args struct {
		ctx                context.Context
		l                  log.Logger
		cli                *client.Client
		id                 string
		failedActionStatus chan pb.State
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

			waitFailedContainer(tArgs.ctx, tArgs.l, tArgs.cli, tArgs.id, tArgs.failedActionStatus)

		})
	}
}

func Test_removeContainer(t *testing.T) {
	type args struct {
		ctx context.Context
		l   log.Logger
		cli *client.Client
		id  string
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			err := removeContainer(tArgs.ctx, tArgs.l, tArgs.cli, tArgs.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("removeContainer error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
