package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type mockCli struct {
	client.ContainerAPIClient
	response containertypes.ContainerCreateCreatedBody
}

func (mcli *mockCli) ContainerCreate(ctx context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, _ *networktypes.NetworkingConfig, _ *specs.Platform, containerName string) (containertypes.ContainerCreateCreatedBody, error) {
	if containerName == "" {
		return mcli.response, fmt.Errorf("Create container failed, container name is invalid")
	}
	return mcli.response, nil
}

func (mcli *mockCli) ContainerStart(ctx context.Context, containerId string, options types.ContainerStartOptions) error {
	if containerId == "" {
		return fmt.Errorf("Container start failed, container id is invalid")
	}
	return nil
}

func (mcli *mockCli) ContainerRemove(ctx context.Context, containerId string, options types.ContainerRemoveOptions) error {
	if containerId == "" {
		return fmt.Errorf("Container remove failed, container id is invalid")
	}
	return nil
}

func (mcli *mockCli) ContainerWait(ctx context.Context, containerId string, condition containertypes.WaitCondition) (<-chan containertypes.ContainerWaitOKBody, <-chan error) {
	status := make(chan containertypes.ContainerWaitOKBody)
	err := make(chan error)
	if containerId == "" {
		go func() {
			err <- fmt.Errorf("Container wait failed, container id is invalid")
		}()

	} else {
		go func() {
			status <- containertypes.ContainerWaitOKBody{StatusCode: 0}
		}()

	}

	return status, err
}

func Test_removeContainer(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx                  context.Context
		cli                  client.ContainerAPIClient
		containerId          string
		forceRemoveContainer bool
		linkRemove           bool
		volumesRemove        bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "container1",
			args:    args{ctx: context.Background(), cli: mock, containerId: "", forceRemoveContainer: true, linkRemove: true, volumesRemove: true},
			wantErr: false,
		},
		{
			name:    "container2",
			args:    args{ctx: context.Background(), cli: mock, containerId: "57cb9052bb5c", forceRemoveContainer: false, linkRemove: false, volumesRemove: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := removeContainer(tt.args.ctx, tt.args.cli, tt.args.containerId, tt.args.forceRemoveContainer, tt.args.linkRemove, tt.args.volumesRemove); (err != nil) != tt.wantErr {
				t.Errorf("removeContainer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_startContainer(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx         context.Context
		cli         client.ContainerAPIClient
		containerId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "container1", args: args{ctx: context.Background(), cli: mock, containerId: ""}, wantErr: false,
		},

		{
			name: "container2", args: args{ctx: context.Background(), cli: mock, containerId: "57cb9052bb5c"}, wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := startContainer(tt.args.ctx, tt.args.cli, tt.args.containerId)
			if (err != nil) != tt.wantErr {
				t.Errorf("startContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_createContainer(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx         context.Context
		mcli        client.ContainerAPIClient
		cActions    containerAction
		cmd         []string
		registry    string
		imageName   string
		captureLogs bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "container_testing1",
			args: args{ctx: context.Background(),
				mcli:        mock,
				cActions:    containerAction{},
				cmd:         []string{"testing containers1"},
				registry:    "registry",
				imageName:   "image testing1",
				captureLogs: false},
			wantErr: false,
		},
		{
			name: "container_testing2",
			args: args{ctx: context.Background(),
				mcli:        mock,
				cActions:    containerAction{},
				cmd:         []string{"testing containers2"},
				registry:    "registry2",
				imageName:   "image testing2",
				captureLogs: false},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createContainer(tt.args.ctx, tt.args.mcli, tt.args.cActions, tt.args.cmd, tt.args.registry, tt.args.imageName, tt.args.captureLogs)
			t.Logf("%+v, %+v", tt.args, err)

			if (err != nil) != tt.wantErr {
				t.Errorf("createContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func Test_waitContainer(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx         context.Context
		mcli        client.ContainerAPIClient
		containerId string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "testing_waitcontainer1",
			args: args{ctx: context.Background(),
				mcli:        mock,
				containerId: "",
			},
			want:    "FAILED",
			wantErr: false,
		},
		{
			name: "testing_waitcontainer2",
			args: args{ctx: context.Background(),
				mcli:        mock,
				containerId: "57cb9052bb5c",
			},
			want:    "SUCCESS",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := waitContainer(tt.args.ctx, tt.args.mcli, tt.args.containerId)
			if (err != nil) != tt.wantErr {
				t.Errorf("waitContainer() container state = %v, Error = %v", state, err)
			}
		})
	}
}
