//nolint:thelper // misuse of test helpers requires a large refactor into subtests
package db_test

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/workflow"
	"gopkg.in/yaml.v2"
)

func TestCreateTemplate(t *testing.T) {
	ctx := context.Background()

	table := []struct {
		// Name identifies the single test in a table test scenario
		Name string
		// Input is a list of workflows that will be used to pre-populate the database
		Input []*workflow.Workflow
		// InputAsync if set to true inserts all the input concurrently
		InputAsync bool
		// Expectation is the function used to apply the assertions.
		// You can use it to validate if the Input are created as you expect
		Expectation func(*testing.T, []*workflow.Workflow, *db.TinkDB)
		// ExpectedErr is used to check for error during
		// CreateTemplate execution. If you expect a particular error
		// and you want to assert it, you can use this function
		ExpectedErr func(*testing.T, error)
	}{
		{
			Name: "happy-path-single-create-template",
			Input: []*workflow.Workflow{
				func() *workflow.Workflow {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = "545f7ce9-5313-49c6-8704-0ed98814f1f7"
					return w
				}(),
			},
			Expectation: func(t *testing.T, input []*workflow.Workflow, tinkDB *db.TinkDB) {
				wtmpl, err := tinkDB.GetTemplate(ctx, map[string]string{"id": input[0].ID}, false)
				if err != nil {
					t.Error(err)
				}
				w := workflow.MustParse([]byte(wtmpl.GetData()))
				w.ID = wtmpl.GetId()
				w.Name = wtmpl.GetName()
				if dif := cmp.Diff(input[0], w); dif != "" {
					t.Errorf(dif)
				}
			},
		},
		{
			Name: "create-two-template-same-name",
			Input: []*workflow.Workflow{
				func() *workflow.Workflow {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = "545f7ce9-5313-49c6-8704-0ed98814f1f7"
					return w
				}(),
				func() *workflow.Workflow {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = "aaaaaaaa-5313-49c6-8704-bbbbbbbbbbbb"
					return w
				}(),
			},
			ExpectedErr: func(t *testing.T, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"uidx_template_name\"") {
					t.Errorf("\nexpected err: %s\ngot: %s", "pq: duplicate key value violates unique constraint \"uidx_template_name\"", err)
				}
			},
		},
		{
			Name: "update-on-create",
			Input: []*workflow.Workflow{
				func() *workflow.Workflow {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = "545f7ce9-5313-49c6-8704-0ed98814f1f7"
					return w
				}(),
				func() *workflow.Workflow {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = "545f7ce9-5313-49c6-8704-0ed98814f1f7"
					w.Name = "updated-name"
					return w
				}(),
			},
			Expectation: func(t *testing.T, input []*workflow.Workflow, tinkDB *db.TinkDB) {
				wtmpl, err := tinkDB.GetTemplate(context.Background(), map[string]string{"id": input[0].ID}, false)
				if err != nil {
					t.Error(err)
				}
				if wtmpl.GetName() != "updated-name" {
					t.Errorf("expected name to be \"%s\", got \"%s\"", "updated-name", wtmpl.GetName())
				}
			},
		},
		{
			Name:       "create-stress-test",
			InputAsync: true,
			Input: func() []*workflow.Workflow {
				input := []*workflow.Workflow{}
				for ii := 0; ii < 20; ii++ {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = uuid.New().String()
					w.Name = fmt.Sprintf("id_%d", rand.Int())
					t.Log(w.Name)
					input = append(input, w)
				}
				return input
			}(),
			ExpectedErr: func(t *testing.T, err error) {
				if err != nil {
					t.Error(err)
				}
			},
			Expectation: func(t *testing.T, input []*workflow.Workflow, tinkDB *db.TinkDB) {
				count := 0
				err := tinkDB.ListTemplates("%", func(id, n string, in, del *timestamp.Timestamp) error {
					count++
					return nil
				})
				if err != nil {
					t.Error(err)
				}
				if len(input) != count {
					t.Errorf("expected %d templates stored in the database but we got %d", len(input), count)
				}
			},
		},
	}

	for _, s := range table {
		t.Run(s.Name, func(t *testing.T) {
			t.Parallel()
			_, tinkDB, cl := NewPostgresDatabaseClient(ctx, t, NewPostgresDatabaseRequest{
				ApplyMigration: true,
			})
			defer func() {
				err := cl()
				if err != nil {
					t.Error(err)
				}
			}()
			var wg sync.WaitGroup
			wg.Add(len(s.Input))
			for _, tt := range s.Input {
				if s.InputAsync {
					go func(ctx context.Context, tinkDB *db.TinkDB, tt *workflow.Workflow) {
						defer wg.Done()
						err := createTemplateFromWorkflowType(ctx, tinkDB, tt)
						if err != nil {
							s.ExpectedErr(t, err)
						}
					}(ctx, tinkDB, tt)
				} else {
					wg.Done()
					err := createTemplateFromWorkflowType(ctx, tinkDB, tt)
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

func TestCreateTemplate_TwoTemplateWithSameNameButFirstOneIsDeleted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, tinkDB, cl := NewPostgresDatabaseClient(ctx, t, NewPostgresDatabaseRequest{
		ApplyMigration: true,
	})
	defer func() {
		err := cl()
		if err != nil {
			t.Error(err)
		}
	}()

	w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
	w.ID = "545f7ce9-5313-49c6-8704-0ed98814f1f7"
	err := createTemplateFromWorkflowType(ctx, tinkDB, w)
	if err != nil {
		t.Error(err)
	}
	err = tinkDB.DeleteTemplate(ctx, w.ID)
	if err != nil {
		t.Error(err)
	}

	ww := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
	ww.ID = "1111aaaa-5313-49c6-8704-222222aaaaaa"
	err = createTemplateFromWorkflowType(ctx, tinkDB, ww)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteTemplate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, tinkDB, cl := NewPostgresDatabaseClient(ctx, t, NewPostgresDatabaseRequest{
		ApplyMigration: true,
	})
	defer func() {
		err := cl()
		if err != nil {
			t.Error(err)
		}
	}()

	w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
	w.ID = uuid.New().String()
	w.Name = fmt.Sprintf("id_%d", rand.Int())

	err := createTemplateFromWorkflowType(ctx, tinkDB, w)
	if err != nil {
		t.Error(err)
	}
	err = tinkDB.DeleteTemplate(ctx, w.ID)
	if err != nil {
		t.Error(err)
	}

	count := 0
	err = tinkDB.ListTemplates("%", func(id, n string, in, del *timestamp.Timestamp) error {
		count++
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if count != 0 {
		t.Errorf("expected 0 templates stored in the database after delete, but we got %d", count)
	}
}

func TestGetTemplate(t *testing.T) {
	ctx := context.Background()
	expectation := func(t *testing.T, input *workflow.Workflow, tinkDB *db.TinkDB) {
		wtmpl, err := tinkDB.GetTemplate(ctx, map[string]string{"id": input.ID}, false)
		if err != nil {
			t.Error(err)
		}
		w := workflow.MustParse([]byte(wtmpl.GetData()))
		w.ID = wtmpl.GetId()
		w.Name = wtmpl.GetName()
		if dif := cmp.Diff(input, w); dif != "" {
			t.Errorf(dif)
		}
	}
	tests := []struct {
		// Name identifies the single test in a table test scenario
		Name string
		// Input is a list of workflows that will be used to pre-populate the database
		Input []*workflow.Workflow
		// GetAsync if set to true gets all the templates concurrently
		GetAsync bool
		// Expectation is the function used to apply the assertions.
		// You can use it to validate if you get template you expected
		Expectation func(*testing.T, *workflow.Workflow, *db.TinkDB)
	}{
		{
			Name: "get-template",
			Input: []*workflow.Workflow{
				func() *workflow.Workflow {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = "545f7ce9-5313-49c6-8704-0ed98814f1f7"
					return w
				}(),
			},
			Expectation: expectation,
		},
		{
			Name:     "stress-get-template",
			GetAsync: true,
			Input: func() []*workflow.Workflow {
				input := []*workflow.Workflow{}
				for i := 0; i < 20; i++ {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = uuid.New().String()
					w.Name = fmt.Sprintf("id_%d", rand.Int())
					t.Log(w.Name)
					input = append(input, w)
				}
				return input
			}(),
			Expectation: expectation,
		},
	}
	for _, s := range tests {
		t.Run(s.Name, func(t *testing.T) {
			t.Parallel()
			_, tinkDB, cl := NewPostgresDatabaseClient(ctx, t, NewPostgresDatabaseRequest{
				ApplyMigration: true,
			})
			defer func() {
				err := cl()
				if err != nil {
					t.Error(err)
				}
			}()

			for _, in := range s.Input {
				err := createTemplateFromWorkflowType(ctx, tinkDB, in)
				if err != nil {
					t.Error(err)
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(s.Input))
			for _, in := range s.Input {
				if s.GetAsync {
					go func(t *testing.T, wf *workflow.Workflow, db *db.TinkDB) {
						defer wg.Done()
						s.Expectation(t, wf, db)
					}(t, in, tinkDB)
				} else {
					wg.Done()
					s.Expectation(t, in, tinkDB)
				}
			}
			wg.Wait()
		})
	}
}

func TestGetTemplateWithInvalidID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, tinkDB, cl := NewPostgresDatabaseClient(ctx, t, NewPostgresDatabaseRequest{
		ApplyMigration: true,
	})
	defer func() {
		err := cl()
		if err != nil {
			t.Error(err)
		}
	}()

	id := uuid.New().String()
	_, err := tinkDB.GetTemplate(ctx, map[string]string{"id": id}, false)
	if err == nil {
		t.Error("expected error, got nil")
	}

	want := "no rows in result set"
	// TODO: replace with errors.Is
	if !strings.Contains(err.Error(), want) {
		t.Error(fmt.Errorf("unexpected output, looking for %q as a substring in %q", want, err.Error()))
	}
}

func createTemplateFromWorkflowType(ctx context.Context, tinkDB *db.TinkDB, tt *workflow.Workflow) error {
	uID := uuid.MustParse(tt.ID)
	content, err := yaml.Marshal(tt)
	if err != nil {
		return err
	}
	err = tinkDB.CreateTemplate(ctx, tt.Name, string(content), uID)
	if err != nil {
		return err
	}
	return nil
}
