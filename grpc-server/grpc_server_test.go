package grpcserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/grpc"
	"github.com/packethost/pkg/log"
)

func TestSetupGRPC(t *testing.T) {
	type input struct {
		server string
		client string
	}
	tests := map[string]struct {
		input input
		want  []string
		err   error
	}{
		"successful grpc client call":     {input: input{server: "127.0.0.1:55005", client: "127.0.0.1:55005"}, want: []string{"HardwareService", "TemplateService", "WorkflowService", "ServerReflection"}},
		"grpc client fail to communicate": {input: input{server: "127.0.0.1:0", client: "127.0.0.1:55007"}, err: fmt.Errorf("failed to list services from reflection enabled gRPC server: rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing dial tcp 127.0.0.1:55007: connect: connection refused\"")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			errCh := make(chan error)
			logger, _ := log.Init("test_package")
			SetupGRPC(ctx, logger, &ConfigGRPCServer{
				Facility:      "onprem",
				TLSCert:       "just can't be an empty string",
				GRPCAuthority: tc.input.server,
			}, errCh)
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
