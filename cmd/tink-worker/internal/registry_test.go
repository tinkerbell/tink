package internal

import (
	"context"
	"reflect"
	"testing"

	"github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
)

func TestNewRegistryConnDetails(t *testing.T) {
	type args struct {
		registry string
		user     string
		pwd      string
		logger   log.Logger
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 *RegistryConnDetails
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := NewRegistryConnDetails(tArgs.registry, tArgs.user, tArgs.pwd, tArgs.logger)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewRegistryConnDetails got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestRegistryConnDetails_NewClient(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *RegistryConnDetails
		inspect func(r *RegistryConnDetails, t *testing.T) //inspects receiver after test run

		want1      *client.Client
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			got1, err := receiver.NewClient()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("RegistryConnDetails.NewClient got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("RegistryConnDetails.NewClient error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func TestRegistryConnDetails_pullImage(t *testing.T) {
	type args struct {
		ctx   context.Context
		cli   *client.Client
		image string
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *RegistryConnDetails
		inspect func(r *RegistryConnDetails, t *testing.T) //inspects receiver after test run

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
			err := receiver.pullImage(tArgs.ctx, tArgs.cli, tArgs.image)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("RegistryConnDetails.pullImage error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
