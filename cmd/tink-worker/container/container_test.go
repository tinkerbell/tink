package container

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/go-cmp/cmp"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type mockCli struct {
	client.ContainerAPIClient
	response containertypes.ContainerCreateCreatedBody
	inspect  types.ContainerJSON
}

func (mcli *mockCli) ContainerCreate(_ context.Context, _ *containertypes.Config, _ *containertypes.HostConfig, _ *networktypes.NetworkingConfig, _ *specs.Platform, _ string) (containertypes.ContainerCreateCreatedBody, error) {
	return mcli.response, nil
}

func (mcli *mockCli) ContainerInspect(_ context.Context, _ string) (types.ContainerJSON, error) {
	return mcli.inspect, nil
}

func (mcli *mockCli) ContainerRemove(_ context.Context, _ string, _ types.ContainerRemoveOptions) error {
	return nil
}

func (mcli *mockCli) ContainerStart(_ context.Context, _ string, _ types.ContainerStartOptions) error {
	return nil
}

func (mcli *mockCli) ContainerWait(_ context.Context, id string, _ containertypes.WaitCondition) (<-chan containertypes.ContainerWaitOKBody, <-chan error) {
	status := make(chan containertypes.ContainerWaitOKBody)
	err := make(chan error)
	if id == "" {
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

func TestRemoveContainer(t *testing.T) {
	mock := &mockCli{}

	type args struct {
		ctx  context.Context
		cli  client.ContainerAPIClient
		id   string
		opts RemoveOpts
	}
	tests := []struct {
		name    string
		data    args
		wantErr error
		c       Cdata
	}{
		{
			name:    "remove_container_with_no_containerID",
			data:    args{ctx: context.Background(), cli: mock, id: "", opts: RemoveOpts{Force: true, Links: true, Volumes: true}},
			wantErr: errors.New("empty string is not a valid id"),
		},
		{
			name:    "remove_container_with_containerID",
			data:    args{ctx: context.Background(), cli: mock, id: "57cb9052bb5c", opts: RemoveOpts{Force: false, Links: false, Volumes: false}},
			wantErr: nil,
		},
		{
			name:    "remove_container_with_no_client",
			data:    args{ctx: context.Background(), cli: nil, id: "57cb9052bb5c", opts: RemoveOpts{Force: true, Links: true, Volumes: true}},
			wantErr: errors.New("ContainerAPIClient interface is not valid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Remove(tt.data.ctx, tt.data.cli, tt.data.id, tt.data.opts)
			if err != nil {
				diff := cmp.Diff(tt.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestStartContainer(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx context.Context
		cli client.ContainerAPIClient
		id  string
	}
	tests := []struct {
		name    string
		data    args
		wantErr error
		c       Cdata
	}{
		{
			name: "start_container_with_no_containerID", data: args{ctx: context.Background(), cli: mock, id: ""}, wantErr: errors.New("empty string is not a valid id"),
		},

		{
			name: "start_container_with_containerID", data: args{ctx: context.Background(), cli: mock, id: "57cb9052bb5c"}, wantErr: nil,
		},
		{
			name: "start_container_with_no_client", data: args{ctx: context.Background(), cli: nil, id: "57cb9052bb5c"}, wantErr: errors.New("ContainerAPIClient interface is not valid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Start(tt.data.ctx, tt.data.cli, tt.data.id)
			if err != nil {
				diff := cmp.Diff(tt.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestCreateContainer(t *testing.T) {
	mock := &mockCli{}

	type args struct {
		ctx  context.Context
		mcli client.ContainerAPIClient
	}
	tests := []struct {
		name    string
		data    args
		wantErr error
		c       Cdata
	}{
		{
			name: "create_container_with_no_container_name",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
			},
			wantErr: errors.New("container name and image name are required data"),
			c: Cdata{
				Data: ConfigData{
					Pid:            "56789",
					Env:            []string{""},
					Name:           "",
					VolumeBindings: []string{"/dev:/dev"},
					Cmd:            []string{"Docker"},
					Registry:       "registry",
					ImageName:      "image testing1",
					CaptureLogs:    false,
				},
			},
		},
		{
			name: "create_container_with_container_name",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
			},
			wantErr: nil,
			c: Cdata{
				Data: ConfigData{
					Pid:            "12567895",
					Env:            []string{""},
					Name:           "test2",
					VolumeBindings: []string{"/dev:/dev-dev1"},
					Cmd:            []string{"Docker"},
					Registry:       "registry",
					ImageName:      "image testing2",
					CaptureLogs:    false,
				},
			},
		},
		{
			name: "create_container_with_not_valid_volume",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
			},
			wantErr: errors.New("volume is invalid format. valid format:'/dev:/dev'"),
			c: Cdata{
				Data: ConfigData{
					Pid:            "12568950",
					Env:            []string{""},
					Name:           "test3",
					VolumeBindings: []string{"/dev:/dev$dev1"},
					Cmd:            []string{"Docker"},
					Registry:       "registry",
					ImageName:      "image testing2",
					CaptureLogs:    false,
				},
			},
		},
		{
			name: "create_container_with_valid_volume",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
			},
			wantErr: nil,
			c: Cdata{
				Data: ConfigData{
					Pid:            "12568950",
					Env:            []string{""},
					Name:           "test4",
					VolumeBindings: []string{"/dev:/dev"},
					Cmd:            []string{"Docker"},
					Registry:       "registry",
					ImageName:      "image testing2",
					CaptureLogs:    false,
				},
			},
		},
		{
			name: "create_container_with_no_cmd",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
			},
			wantErr: errors.New("cmd has invalid data"),
			c: Cdata{
				Data: ConfigData{
					Pid:            "12568950",
					Env:            []string{""},
					Name:           "test5",
					VolumeBindings: []string{"/dev:/dev-d"},
					Cmd:            []string{},
					Registry:       "registry",
					ImageName:      "image testing2",
					CaptureLogs:    false,
				},
			},
		},
		{
			name: "create_container_with_cmd",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
			},
			wantErr: nil,
			c: Cdata{
				Data: ConfigData{
					Pid:            "12568950",
					Env:            []string{""},
					Name:           "test6",
					VolumeBindings: []string{"/dev:/dev"},
					Cmd:            []string{"Docker"},
					Registry:       "registry",
					ImageName:      "image testing",
					CaptureLogs:    false,
				},
			},
		},
		{
			name: "create_container_with_no_image",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
			},
			wantErr: errors.New("container name and image name are required data"),
			c: Cdata{
				Data: ConfigData{
					Pid:            "12568950",
					Env:            []string{""},
					Name:           "test7",
					VolumeBindings: []string{"/dev:/dev"},
					Cmd:            []string{"testing container"},
					Registry:       "registry",
					ImageName:      "",
					CaptureLogs:    false,
				},
			},
		},
		{
			name: "create_container_with_image",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
			},
			wantErr: nil,
			c: Cdata{
				Data: ConfigData{
					Pid:            "12568950",
					Env:            []string{""},
					Name:           "test8",
					VolumeBindings: []string{"/dev:/dev"},
					Cmd:            []string{"testing container"},
					Registry:       "registry",
					ImageName:      "image testing",
					CaptureLogs:    false,
				},
			},
		},
		{
			name: "create_container_with_no_client",
			data: args{
				ctx:  context.Background(),
				mcli: nil,
			},
			wantErr: errors.New("ContainerAPIClient interface is not valid"),
			c: Cdata{
				Data: ConfigData{
					Pid:            "12568950",
					Env:            []string{""},
					Name:           "test7",
					VolumeBindings: []string{"/dev:/dev"},
					Cmd:            []string{"testing container"},
					Registry:       "registry",
					ImageName:      "testing image",
					CaptureLogs:    false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.c.Create(tt.data.ctx, tt.data.mcli)
			if err != nil {
				diff := cmp.Diff(tt.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestWaitContainer(t *testing.T) {
	mock := &mockCli{}
	type args struct {
		ctx  context.Context
		mcli client.ContainerAPIClient
		id   string
	}
	tests := []struct {
		name    string
		data    args
		want    string
		wantErr error
		c       Cdata
	}{
		{
			name: "wait_container_with_no_containerID",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
				id:   "",
			},
			wantErr: errors.New("empty string is not a valid id"),
		},
		{
			name: "wait_container_with_containerID",
			data: args{
				ctx:  context.Background(),
				mcli: mock,
				id:   "57cb9052bb5c",
			},
			wantErr: nil,
		},
		{
			name: "wait_container_with_no_client",
			data: args{
				ctx:  context.Background(),
				mcli: nil,
				id:   "57cb9052bb5c",
			},
			wantErr: errors.New("ContainerAPIClient interface is not valid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.c.Wait(tt.data.ctx, tt.data.mcli, tt.data.id)
			if err != nil {
				diff := cmp.Diff(tt.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
