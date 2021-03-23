package cmd

import (
	"reflect"
	"testing"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func TestNewRootCommand(t *testing.T) {
	type args struct {
		version string
		logger  log.Logger
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 *cobra.Command
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := NewRootCommand(tArgs.version, tArgs.logger)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewRootCommand got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func Test_createViper(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1      *viper.Viper
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1, err := createViper(tArgs.logger)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("createViper got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("createViper error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_applyViper(t *testing.T) {
	type args struct {
		v   *viper.Viper
		cmd *cobra.Command
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

			err := applyViper(tArgs.v, tArgs.cmd)

			if (err != nil) != tt.wantErr {
				t.Fatalf("applyViper error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_tryClientConnection(t *testing.T) {
	type args struct {
		logger        log.Logger
		retryInterval time.Duration
		retries       int
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1      *grpc.ClientConn
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		//TODO: Add test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1, err := tryClientConnection(tArgs.logger, tArgs.retryInterval, tArgs.retries)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("tryClientConnection got1 = %v, want1: %v", got1, tt.want1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("tryClientConnection error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
