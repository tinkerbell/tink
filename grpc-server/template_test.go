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

	noTimeoutTemplate = `version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "Invalid Template"
    worker: "{{.device_3}}"
    actions:
    - name: "action_without_timeout"
      image: hello-world`
)

func TestCreateTemplate(t *testing.T) {
	type (
		args struct {
			db        mock.DB
			name      []string
			templates []string
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
				db:        mock.DB{},
				name:      []string{"template_1"},
				templates: []string{template1},
			},
			want: want{
				expectedError: false,
			},
		},

		"SuccessfullMultipleTemplateCreation": {
			args: args{
				db:        mock.DB{},
				name:      []string{"template_1", "template_2"},
				templates: []string{template1, template2},
			},
			want: want{
				expectedError: false,
			},
		},

		"FailedMultipleTemplateCreationWithSameName": {
			args: args{
				db:        mock.DB{},
				name:      []string{"template_1", "template_1"},
				templates: []string{template1, template2},
			},
			want: want{
				expectedError: true,
			},
		},

		"TemplateWithNoTimeout": {
			args: args{
				db:        mock.DB{},
				name:      []string{"noTimeoutTemplate"},
				templates: []string{noTimeoutTemplate},
			},
			want: want{
				expectedError: true,
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(tc.args.db)
			tc.args.db.ClearTemplateDB()
			index := 0
			res, err := s.CreateTemplate(context.TODO(), &pb.WorkflowTemplate{Name: tc.args.name[index], Data: tc.args.templates[index]})
			if name == "TemplateWithNoTimeout" {
				assert.True(t, tc.want.expectedError)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, res)
			if err == nil && len(tc.args.templates) > 1 {
				index++
				_, err = s.CreateTemplate(context.TODO(), &pb.WorkflowTemplate{Name: tc.args.name[index], Data: tc.args.templates[index]})
			} else {
				return
			}
			if err != nil {
				assert.Error(t, err)
				assert.True(t, tc.want.expectedError)
			} else {
				assert.Nil(t, err)
				assert.False(t, tc.want.expectedError)
			}
		})
	}
}
