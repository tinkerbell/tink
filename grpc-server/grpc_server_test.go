package grpcserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/grpc"
	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/server"
)

func TestSetupGRPC(t *testing.T) {
	type input struct {
		server string
		client string
	}
	tests := []struct {
		name  string
		input input
		want  []string
		err   error
	}{
		{
			name: "successful grpc client call",
			input: input{
				server: "127.0.0.1:55005",
				client: "127.0.0.1:55005",
			},
			want: []string{"HardwareService", "TemplateService", "WorkflowService", "ServerReflection"},
		},
		{
			name: "grpc client fail to communicate",
			input: input{
				server: "127.0.0.1:0",
				client: "127.0.0.1:55007",
			},
			err: fmt.Errorf("failed to list services from reflection enabled gRPC server: rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing dial tcp 127.0.0.1:55007: connect: connection refused\""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			errCh := make(chan error)
			logger := log.Test(t, "test_package")
			tinkServer, _ := server.NewDBServer(
				logger,
				nil,
			)
			_, err := SetupGRPC(
				ctx,
				tinkServer,
				tc.input.server,
				nil,
				errCh)
			if err != nil {
				t.Errorf("failed to set up gRPC server: %v", err)
				return
			}

			client, err := grpc.NewClient(tc.input.client, "name", true, false, "", "", "", nil)
			if err != nil {
				t.Fatal(err)
			}
			var retries int
		RETRY:
			pkgs, err := client.ListPackages()
			if err != nil {
				// there's a timing issue with the grpc server, so we retry
				if retries != 2 && tc.err == nil {
					retries++
					time.Sleep(1 * time.Second)
					goto RETRY
				}
				if tc.err != nil {
					if diff := cmp.Diff(tc.err.Error(), err.Error()); diff != "" {
						t.Error(diff)
					}
				} else {
					t.Errorf("got unexpected error: %v", err)
				}
			} else {
				var got []string
				for _, proto := range pkgs {
					for _, elem := range proto.AsFileDescriptorProto().GetService() {
						got = append(got, elem.GetName())
					}
				}
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func TestGetCerts(t *testing.T) {
	cases := []struct {
		name      string
		setupFunc func(t *testing.T) (string, error)
		wanterr   error
	}{
		{
			"Real key file",
			func(t *testing.T) (string, error) {
				t.Helper()
				return "./testdata", nil
			},
			nil,
		},
		{
			"No cert",
			func(t *testing.T) (string, error) {
				t.Helper()
				return "./not-a-directory", nil
			},
			fmt.Errorf("failed to load TLS files: open not-a-directory/bundle.pem: no such file or directory"),
		},
		{
			"empty content",
			func(t *testing.T) (string, error) {
				t.Helper()
				tdir := t.TempDir()
				err := os.WriteFile(filepath.Join(tdir, "bundle.pem"), []byte{}, 0o644)
				if err != nil {
					return "", err
				}
				err = os.WriteFile(filepath.Join(tdir, "server-key.pem"), []byte{}, 0o644)
				if err != nil {
					return "", err
				}
				return tdir, nil
			},
			fmt.Errorf("failed to load TLS files: tls: failed to find any PEM data in certificate input"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input, err := tc.setupFunc(t)
			if err != nil {
				t.Errorf("Failed to setup test: %v", err)
				return
			}
			gotCert, err := GetCerts(input)

			if tc.wanterr == nil {
				if gotCert == nil {
					t.Error("Missing expected cert, got nil")
				}
			}
			if tc.wanterr == nil && err == nil {
				return
			}
			if tc.wanterr != nil {
				if err == nil {
					t.Errorf("Missing expected error %s", tc.wanterr.Error())
					return
				}
				if tc.wanterr.Error() != err.Error() {
					t.Errorf("Got different error.\nWanted:\n  %s\nGot:\n  %s", tc.wanterr.Error(), err.Error())
				}
				return
			}
		})
	}
}
