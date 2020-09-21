package grpcserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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

type templates struct {
	id   uuid.UUID
	data string
}

var templateDB = map[string]interface{}{}

// ClearTemplateDB clear all the templates
func clearTemplateDB() {
	templateDB = map[string]interface{}{}
}

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
				db: mock.DB{
					CreateTemplateFunc: func(ctx context.Context, name string, data string, id uuid.UUID) error {
						templateDB[name] = templates{
							id:   id,
							data: data,
						}
						return nil
					},
				},
				name:      []string{"template_1"},
				templates: []string{template1},
			},
			want: want{
				expectedError: false,
			},
		},

		"SuccessfullMultipleTemplateCreation": {
			args: args{
				db: mock.DB{
					CreateTemplateFunc: func(ctx context.Context, name string, data string, id uuid.UUID) error {
						if len(templateDB) > 0 {
							if _, ok := templateDB[name]; ok {
								return fmt.Errorf("Template name already exist in the database")
							}
							templateDB[name] = templates{
								id:   id,
								data: data,
							}
							return nil

						}
						templateDB[name] = templates{
							id:   id,
							data: data,
						}
						return nil
					},
				},
				name:      []string{"template_1", "template_2"},
				templates: []string{template1, template2},
			},
			want: want{
				expectedError: false,
			},
		},

		"FailedMultipleTemplateCreationWithSameName": {
			args: args{
				db: mock.DB{
					CreateTemplateFunc: func(ctx context.Context, name string, data string, id uuid.UUID) error {
						if len(templateDB) > 0 {
							if _, ok := templateDB[name]; ok {
								return errors.New("Template name already exist in the database")
							}
							templateDB[name] = templates{
								id:   id,
								data: data,
							}
							return nil

						}
						templateDB[name] = templates{
							id:   id,
							data: data,
						}
						return nil
					},
				},
				name:      []string{"template_1", "template_1"},
				templates: []string{template1, template2},
			},
			want: want{
				expectedError: true,
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := testServer(tc.args.db)
			clearTemplateDB()
			index := 0
			res, err := s.CreateTemplate(context.TODO(), &pb.WorkflowTemplate{Name: tc.args.name[index], Data: tc.args.templates[index]})
			assert.Nil(t, err)
			assert.NotNil(t, res)
			if err == nil && len(tc.args.templates) > 1 {
				index++
				res, err = s.CreateTemplate(context.TODO(), &pb.WorkflowTemplate{Name: tc.args.name[index], Data: tc.args.templates[index]})
			} else {
				return
			}
			if err != nil {
				assert.Error(t, err)
				assert.Empty(t, res)
				assert.True(t, tc.want.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, res)
				assert.False(t, tc.want.expectedError)
			}
		})
	}
}
