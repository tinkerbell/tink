package db_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/protos/hardware"
	pb "github.com/tinkerbell/tink/protos/workflow"

	"github.com/tinkerbell/tink/workflow"
)

type input struct {
	devices       string
	template      *workflow.Workflow
	hardware      *hardware.Hardware
	workflowCount int
}

func TestCreateWorkflow(t *testing.T) {
	tests := []struct {
		// Name identifies the single test in a table test scenario
		Name string
		// InputAsync if set to true inserts all the input concurrently
		InputAsync bool
		// Input is a struct that will be used to create a workflow and pre-populate the database
		Input *input
		// Expectation is the function used to apply the assertions.
		// You can use it to validate if the Input are created as you expect
		Expectation func(t *testing.T, in *input, tinkDB *db.TinkDB)
		// ExpectedErr is used to check for error during
		// CreateWorkflow execution. If you expect a particular error
		// and you want to assert it, you can use this function
		ExpectedErr func(*testing.T, error)
	}{
		{
			Name: "create-single-workflow",
			Input: &input{
				workflowCount: 1,
				devices:       "{\"device_1\":\"08:00:27:00:00:01\"}",
				hardware:      readHardwareData("./testdata/hardware.json"),
				template: func() *workflow.Workflow {
					tmp := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					tmp.ID = uuid.New().String()
					tmp.Name = fmt.Sprintf("id_%d", rand.Int())
					return tmp
				}(),
			},
			Expectation: func(t *testing.T, in *input, tinkDB *db.TinkDB) {
				count := 0
				err := tinkDB.ListWorkflows(func(wf db.Workflow) error {
					count = count + 1
					return nil
				})
				if err != nil {
					t.Error(err)
				}
				if count != in.workflowCount {
					t.Errorf("expected %d workflows stored in the database but we got %d", in.workflowCount, count)
				}
			},
		},
		{
			Name: "create-fails-invalid-worker-address",
			Input: &input{
				workflowCount: 0,
				devices:       "{\"invalid_device\":\"08:00:27:00:00:01\"}",
				hardware:      readHardwareData("./testdata/hardware.json"),
				template: func() *workflow.Workflow {
					tmp := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					tmp.ID = uuid.New().String()
					tmp.Name = fmt.Sprintf("id_%d", rand.Int())
					return tmp
				}(),
			},
			Expectation: func(t *testing.T, in *input, tinkDB *db.TinkDB) {
				count := 0
				err := tinkDB.ListWorkflows(func(wf db.Workflow) error {
					count = count + 1
					return nil
				})
				if err != nil {
					t.Error(err)
				}
				if count != in.workflowCount {
					t.Errorf("expected %d workflows stored in the database but we got %d", in.workflowCount, count)
				}
			},
		},
		{
			Name:       "stress-create-workflow",
			InputAsync: true,
			Input: &input{
				workflowCount: 20,
				devices:       "{\"device_1\":\"08:00:27:00:00:01\"}",
				hardware:      readHardwareData("./testdata/hardware.json"),
				template: func() *workflow.Workflow {
					tmp := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					tmp.ID = uuid.New().String()
					tmp.Name = fmt.Sprintf("id_%d", rand.Int())
					return tmp
				}(),
			},
			Expectation: func(t *testing.T, in *input, tinkDB *db.TinkDB) {
				count := 0
				err := tinkDB.ListWorkflows(func(wf db.Workflow) error {
					count = count + 1
					return nil
				})
				if err != nil {
					t.Error(err)
				}
				if count != in.workflowCount {
					t.Errorf("expected %d workflows stored in the database but we got %d", in.workflowCount, count)
				}
			},
		},
	}

	ctx := context.Background()
	for _, s := range tests {
		t.Run(s.Name, func(t *testing.T) {
			t.Parallel()
			_, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{
				ApplyMigration: true,
			})
			defer func() {
				err := cl()
				if err != nil {
					t.Error(err)
				}
			}()

			err := createHardware(ctx, tinkDB, s.Input.hardware)
			if err != nil {
				t.Error(err)
			}
			err = createTemplateFromWorkflowType(ctx, tinkDB, s.Input.template)
			if err != nil {
				t.Error(err)
			}

			var wg sync.WaitGroup
			wg.Add(s.Input.workflowCount)
			for i := 0; i < s.Input.workflowCount; i++ {
				if s.InputAsync {
					go func(ctx context.Context, tinkDB *db.TinkDB, in *input) {
						defer wg.Done()
						_, err := createWorkflow(ctx, tinkDB, in)
						if err != nil {
							s.ExpectedErr(t, err)
						}
					}(ctx, tinkDB, s.Input)
				} else {
					wg.Done()
					_, err := createWorkflow(ctx, tinkDB, s.Input)
					if err != nil {
						s.ExpectedErr(t, err)
					}
				}
			}
			wg.Wait()
			s.Expectation(t, s.Input, tinkDB)
		})
	}
}

func TestDeleteWorkflow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{
		ApplyMigration: true,
	})
	defer func() {
		err := cl()
		if err != nil {
			t.Error(err)
		}
	}()

	in := &input{
		devices:  "{\"device_1\":\"08:00:27:00:00:01\"}",
		hardware: readHardwareData("./testdata/hardware.json"),
		template: func() *workflow.Workflow {
			tmp := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
			tmp.ID = uuid.New().String()
			tmp.Name = fmt.Sprintf("id_%d", rand.Int())
			return tmp
		}(),
	}
	err := createHardware(ctx, tinkDB, in.hardware)
	if err != nil {
		t.Error(err)
	}
	err = createTemplateFromWorkflowType(ctx, tinkDB, in.template)
	if err != nil {
		t.Error(err)
	}

	wfID, err := createWorkflow(ctx, tinkDB, in)
	if err != nil {
		t.Error(err)
	}

	err = tinkDB.DeleteWorkflow(ctx, wfID, pb.State_value[pb.State_STATE_PENDING.String()])
	if err != nil {
		t.Error(err)
	}

	count := 0
	err = tinkDB.ListWorkflows(func(wf db.Workflow) error {
		count = count + 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if count != 0 {
		t.Errorf("expected 0 workflows stored in the database after delete, but we got %d", count)
	}
}

func createWorkflow(ctx context.Context, tinkDB *db.TinkDB, in *input) (string, error) {
	_, _, tmpData, err := tinkDB.GetTemplate(context.Background(), map[string]string{"id": in.template.ID}, false)
	if err != nil {
		return "", err
	}

	data, err := workflow.RenderTemplate(in.template.ID, tmpData, []byte(in.devices))
	if err != nil {
		return "", err
	}

	id := uuid.New()
	wf := db.Workflow{
		ID:       id.String(),
		Template: in.template.ID,
		Hardware: in.devices,
	}
	err = tinkDB.CreateWorkflow(ctx, wf, data, id)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
