package grpcserver

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/db/mock"
	"github.com/tinkerbell/tink/protos/workflow"
)

const (
	templateID   = "e29b6444-1de7-4a69-bf25-6ea4ae869005"
	hw           = `{"device_1": "08:00:27:00:00:01"}`
	templateData = `version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60`
)

func TestCreateWorkflow(t *testing.T) {
	type (
		args struct {
			db                     mock.DB
			wfTemplate, wfHardware string
		}
		want struct {
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"FailedToGetTemplate": {
			args: args{
				db: mock.DB{
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool) (string, string, string, error) {
						return "", "", "", errors.New("failed to get template")
					},
				},
				wfTemplate: templateID,
				wfHardware: hw,
			},
			want: want{
				expectedError: true,
			},
		},
		"FailedCreatingWorkflow": {
			args: args{
				db: mock.DB{
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool) (string, string, string, error) {
						return "", "", templateData, nil
					},
					CreateWorkflowFunc: func(ctx context.Context, wf db.Workflow, data string, id uuid.UUID) error {
						return errors.New("failed to create a workfow")
					},
				},
				wfTemplate: templateID,
				wfHardware: hw,
			},
			want: want{
				expectedError: true,
			},
		},
		"SuccessCreatingWorkflow": {
			args: args{
				db: mock.DB{
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool) (string, string, string, error) {
						return "", "", templateData, nil
					},
					CreateWorkflowFunc: func(ctx context.Context, wf db.Workflow, data string, id uuid.UUID) error {
						return nil
					},
				},
				wfTemplate: templateID,
				wfHardware: hw,
			},
			want: want{
				expectedError: false,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(tc.args.db)
			res, err := s.CreateWorkflow(ctx, &workflow.CreateRequest{
				Hardware: tc.args.wfHardware,
				Template: tc.args.wfTemplate,
			})
			if err != nil {
				assert.Error(t, err)
				assert.Empty(t, res)
				assert.True(t, tc.want.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
			assert.False(t, tc.want.expectedError)
		})
	}
}
