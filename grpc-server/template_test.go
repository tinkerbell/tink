package grpcserver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/db/mock"
	pb "github.com/tinkerbell/tink/protos/template"
)

const (
	templateNotFoundID       = "abstract-beef-beyond-meat-abominations"
	templateNotFoundName     = "my-awesome-mock-name"
	templateNotFoundTemplate = `version: "0.1"
name: not_found_workflow
global_timeout: 600
tasks:
  - name: "not found"
    worker: "{{.device_1}}"
    actions:
    - name: "not_found"
      image: not-found
      timeout: 60`

	templateID1   = "7cd79119-1959-44eb-8b82-bc15bad4888e"
	templateName1 = "template_1"
	template1     = `version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60`

	templateID2   = "20a18ecf-b9f2-4348-8668-52f672d49208"
	templateName2 = "template_2"
	template2     = `version: "0.1"
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
		"SuccessfulTemplateCreation": {
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

		"SuccessfulMultipleTemplateCreation": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						"template_1": mock.Template{
							Data:    template1,
							Deleted: false,
						},
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
						"template_1": mock.Template{
							Data:    template1,
							Deleted: false,
						},
					},
				},
				name:     "template_1",
				template: template2,
			},
			want: want{
				expectedError: true,
			},
		},

		"SuccessfulTemplateCreationAfterDeletingWithSameName": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						"template_1": mock.Template{
							Data:    template1,
							Deleted: true,
						},
					},
				},
				name:     "template_1",
				template: template2,
			},
			want: want{
				expectedError: false,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := testServer(t, tc.args.db)
			res, err := s.CreateTemplate(ctx, &pb.WorkflowTemplate{Name: tc.args.name, Data: tc.args.template})
			if tc.want.expectedError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, res)
			}
			tc.args.db.ClearTemplateDB()
		})
	}
}

func TestGetTemplate(t *testing.T) {
	type (
		args struct {
			db         mock.DB
			getRequest *pb.GetRequest
		}
	)
	testCases := map[string]struct {
		args args
		id   string
		name string
		data string
		err  bool
	}{
		"SuccessfulTemplateGet_Name": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						templateName1: template1,
					},
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool, revision int32) (db.Template, error) {
						if fields["id"] == templateID1 || fields["name"] == templateName1 {
							return db.Template{ID: templateID1, Name: templateName1, Data: template1}, nil
						}
						return db.Template{ID: templateNotFoundID, Name: templateNotFoundName, Data: templateNotFoundTemplate}, errors.New("failed to get template")
					},
				},
				getRequest: &pb.GetRequest{GetBy: &pb.GetRequest_Name{Name: templateName1}},
			},
			id:   templateID1,
			name: templateName1,
			data: template1,
			err:  false,
		},

		"FailedTemplateGet_Name": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						templateName1: template1,
					},
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool, revision int32) (db.Template, error) {
						if fields["id"] == templateID1 || fields["name"] == templateName1 {
							return db.Template{ID: templateID1, Name: templateName1, Data: template1}, nil
						}
						return db.Template{ID: templateNotFoundID, Name: templateNotFoundName, Data: templateNotFoundTemplate}, errors.New("failed to get template")
					},
				},
				getRequest: &pb.GetRequest{GetBy: &pb.GetRequest_Name{Name: templateName2}},
			},
			id:   templateNotFoundID,
			name: templateNotFoundName,
			data: templateNotFoundTemplate,
			err:  true,
		},

		"SuccessfulTemplateGet_ID": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						templateName1: template1,
					},
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool, revision int32) (db.Template, error) {
						if fields["id"] == templateID1 || fields["name"] == templateName1 {
							return db.Template{ID: templateID1, Name: templateName1, Data: template1}, nil
						}
						return db.Template{ID: templateNotFoundID, Name: templateNotFoundName, Data: templateNotFoundTemplate}, errors.New("failed to get template")
					},
				},
				getRequest: &pb.GetRequest{GetBy: &pb.GetRequest_Id{Id: templateID1}},
			},
			id:   templateID1,
			name: templateName1,
			data: template1,
			err:  false,
		},

		"FailedTemplateGet_ID": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						templateName1: template1,
					},
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool, revision int32) (db.Template, error) {
						if fields["id"] == templateID1 || fields["name"] == templateName1 {
							return db.Template{ID: templateID1, Name: templateName1, Data: template1}, nil
						}
						return db.Template{ID: templateNotFoundID, Name: templateNotFoundName, Data: templateNotFoundTemplate}, errors.New("failed to get template")
					},
				},
				getRequest: &pb.GetRequest{GetBy: &pb.GetRequest_Id{Id: templateID2}},
			},
			id:   templateNotFoundID,
			name: templateNotFoundName,
			data: templateNotFoundTemplate,
			err:  true,
		},

		"FailedTemplateGet_EmptyRequest": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						templateName1: template1,
					},
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool, revision int32) (db.Template, error) {
						if fields["id"] == templateID1 || fields["name"] == templateName1 {
							return db.Template{ID: templateID1, Name: templateName1, Data: template1}, nil
						}
						return db.Template{ID: templateNotFoundID, Name: templateNotFoundName, Data: templateNotFoundTemplate}, errors.New("failed to get template")
					},
				},
				getRequest: &pb.GetRequest{},
			},
			id:   templateNotFoundID,
			name: templateNotFoundName,
			data: templateNotFoundTemplate,
			err:  true,
		},

		"FailedTemplateGet_NilRequest": {
			args: args{
				db: mock.DB{
					TemplateDB: map[string]interface{}{
						templateName1: template1,
					},
					GetTemplateFunc: func(ctx context.Context, fields map[string]string, deleted bool, revision int32) (db.Template, error) {
						if fields["id"] == templateID1 || fields["name"] == templateName1 {
							return db.Template{ID: templateID1, Name: templateName1, Data: template1}, nil
						}
						return db.Template{ID: templateNotFoundID, Name: templateNotFoundName, Data: templateNotFoundTemplate}, errors.New("failed to get template")
					},
				},
			},
			id:   templateNotFoundID,
			name: templateNotFoundName,
			data: templateNotFoundTemplate,
			err:  true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := testServer(t, tc.args.db)
			res, err := s.GetTemplate(ctx, tc.args.getRequest)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, res)
			}
			assert.Equal(t, res.Id, tc.id)
			assert.Equal(t, res.Name, tc.name)
			assert.Equal(t, res.Data, tc.data)
		})
	}
}

func TestGetRevision(t *testing.T) {
	type (
		args struct {
			db         mock.DB
			templateID string
			revision   int32
		}
	)
	testCases := map[string]struct {
		args args
		data string
		err  bool
	}{
		"get-revision-success": {
			args: args{
				db: mock.DB{
					GetRevisionFunc: func(ctx context.Context, tID string, r int32) (string, error) {
						if tID == templateID1 {
							return template1, nil
						}
						return templateNotFoundTemplate, errors.New("failed to get template")
					},
				},
				templateID: templateID1,
				revision:   1,
			},
			data: template1,
			err:  false,
		},
		"get-revision-failed": {
			args: args{
				db: mock.DB{
					GetRevisionFunc: func(ctx context.Context, tID string, r int32) (string, error) {
						if tID == templateID1 {
							return template1, nil
						}
						return templateNotFoundTemplate, errors.New("failed to get template")
					},
				},
				templateID: templateNotFoundID,
				revision:   1,
			},
			data: templateNotFoundTemplate,
			err:  true,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := testServer(t, tc.args.db)
			res, err := s.GetRevision(ctx, &pb.GetRevisionRequest{
				TemplateId: tc.args.templateID,
				Revision:   tc.args.revision,
			})
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, res)
			}
			assert.Equal(t, res.Data, tc.data)
		})
	}
}
