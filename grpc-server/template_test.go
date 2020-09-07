package grpcserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/db/mock"
	pb "github.com/tinkerbell/tink/protos/template"
)

const (
	template1 = `version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60`

	template2 = `version: "0.1"
name: hello_world_again_workflow
global_timeout: 600
tasks:
  - name: "hello world again"
    worker: "{{.device_2}}"
    actions:
    - name: "hello_world_again"
      image: hello-world
      timeout: 60`
)

func TestCreateTemplate(t *testing.T) {
	type (
		args struct {
			db       mock.DB
			name     string
			template string
		}
		want struct {
			expectedError bool
		}
	)
	testCases := map[string]struct {
		args args
		want want
	}{
		"SuccessfullTemplateCreation": {
			args: args{
				db: mock.DB{
					TemplateDB: make(map[string]interface{}),
				},
				name:     "template_1",
				template: template1,
			},
			want: want{
				expectedError: false,
			},
		},

		"SuccessfullMultipleTemplateCreation": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						"template_1": template1,
					},
				},
				name:     "template_2",
				template: template2,
			},
			want: want{
				expectedError: false,
			},
		},

		"FailedMultipleTemplateCreationWithSameName": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						"template_1": template1,
					},
				},
				name:     "template_1",
				template: template2,
			},
			want: want{
				expectedError: true,
			},
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := testServer(tc.args.db)
			res, err := s.CreateTemplate(context.TODO(), &pb.WorkflowTemplate{Name: tc.args.name, Data: tc.args.template})
			if tc.want.expectedError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, res)
			}
		})
	}
}
