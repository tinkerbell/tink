package delete

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tinkerbell/tink/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewDeleteCommand(t *testing.T) {
	table := []struct {
		name           string
		expectedOutput string
		args           []string
		opt            Options
	}{
		{
			name:           "happy-path",
			expectedOutput: "Deleted\tbeeb5c79\n",
			args:           []string{"beeb5c79"},
			opt: Options{
				DeleteByID: func(c context.Context, fc *client.FullClient, s string) (interface{}, error) {
					return struct{}{}, nil
				},
			},
		},
		{
			name:           "happy-path-multiple-resources",
			expectedOutput: "Deleted\tbeeb5c79\nDeleted\t14810952\nDeleted\te7a91fe9\n",
			args:           []string{"beeb5c79", "14810952", "e7a91fe9"},
			opt: Options{
				DeleteByID: func(c context.Context, fc *client.FullClient, s string) (interface{}, error) {
					return struct{}{}, nil
				},
			},
		},
		{
			name:           "resource-not-found",
			expectedOutput: "Error\tbeeb5c79\tnot found\n",
			args:           []string{"beeb5c79"},
			opt: Options{
				DeleteByID: func(c context.Context, fc *client.FullClient, s string) (interface{}, error) {
					return struct{}{}, status.Error(codes.NotFound, "")
				},
			},
		},
		{
			name:           "multiple-resources-not-found",
			expectedOutput: "Error\tbeeb5c79\tnot found\nError\t14810952\tnot found\n",
			args:           []string{"beeb5c79", "14810952"},
			opt: Options{
				DeleteByID: func(c context.Context, fc *client.FullClient, s string) (interface{}, error) {
					return struct{}{}, status.Error(codes.NotFound, "")
				},
			},
		},
		{
			name:           "only-one-resource-of-two-was-deleted",
			expectedOutput: "Deleted\tbeeb5c79\nError\t14810952\tnot found\n",
			args:           []string{"beeb5c79", "14810952"},
			opt: Options{
				DeleteByID: func(c context.Context, fc *client.FullClient, s string) (interface{}, error) {
					if s == "beeb5c79" {
						return struct{}{}, nil
					}
					return struct{}{}, status.Error(codes.NotFound, "")
				},
			},
		},
	}

	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {
			stdout := bytes.NewBufferString("")
			test.opt.SetFullClient(&client.FullClient{})
			cmd := NewDeleteCommand(test.opt)
			cmd.SetOut(stdout)
			cmd.SetErr(stdout)
			cmd.SetArgs(test.args)
			err := cmd.Execute()
			if err != nil {
				t.Error(err)
			}
			out, err := ioutil.ReadAll(stdout)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(string(out), test.expectedOutput); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
