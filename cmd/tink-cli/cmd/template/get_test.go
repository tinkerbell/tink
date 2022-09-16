package template

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/internal/clientctx"
	"github.com/tinkerbell/tink/protos/template"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetTemplate(t *testing.T) {
	table := []struct {
		counter          int
		Name             string
		ReturnedTemplate []*template.WorkflowTemplate
		Args             []string
		ExpectedStdout   string
	}{
		{
			Name: "happy-path",
			ReturnedTemplate: []*template.WorkflowTemplate{
				{
					Id:   "template-123",
					Name: "hello-test",
					CreatedAt: func() *timestamppb.Timestamp {
						ti, _ := time.Parse("2006", "2016")
						return timestamppb.New(ti)
					}(),
					UpdatedAt: func() *timestamppb.Timestamp {
						ti, _ := time.Parse("2006", "2016")
						return timestamppb.New(ti)
					}(),
				},
			},
			ExpectedStdout: `+--------------+------------+----------------------+----------------------+
| ID           | NAME       | CREATED AT           | UPDATED AT           |
+--------------+------------+----------------------+----------------------+
| template-123 | hello-test | 2016-01-01T00:00:00Z | 2016-01-01T00:00:00Z |
+--------------+------------+----------------------+----------------------+
`,
		},
	}

	for _, s := range table {
		t.Run(s.Name, func(t *testing.T) {
			cl := &client.FullClient{
				TemplateClient: &template.TemplateServiceClientMock{
					ListTemplatesFunc: func(ctx context.Context, in *template.ListRequest, opts ...grpc.CallOption) (template.TemplateService_ListTemplatesClient, error) {
						return &template.TemplateService_ListTemplatesClientMock{
							RecvFunc: func() (*template.WorkflowTemplate, error) {
								s.counter++
								if s.counter > len(s.ReturnedTemplate) {
									return nil, io.EOF
								}
								return s.ReturnedTemplate[s.counter-1], nil
							},
						}, nil
					},
				},
			}
			stdout := bytes.NewBufferString("")
			opt := NewGetOptions()
			cmd := get.NewGetCommand(opt)
			cmd.SetOut(stdout)
			cmd.SetArgs(s.Args)
			err := cmd.ExecuteContext(clientctx.Set(context.Background(), cl))
			if err != nil {
				t.Error(err)
			}
			out, err := io.ReadAll(stdout)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(string(out), s.ExpectedStdout); diff != "" {
				t.Error(diff)
			}
		})
	}
}
