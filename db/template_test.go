package db_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/db/migration"
	"github.com/tinkerbell/tink/workflow"
	"gopkg.in/yaml.v2"

	migrate "github.com/rubenv/sql-migrate"
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
				wID, wName, wData, err := tinkDB.GetTemplate(ctx, map[string]string{"id": input[0].ID}, false)
				if err != nil {
					t.Error(err)
				}
				w := workflow.MustParse([]byte(wData))
				w.ID = wID
				w.Name = wName
				if dif := cmp.Diff(input[0], w); dif != "" {
					t.Errorf(dif)
				}

				count := 0
				err = tinkDB.ListRevisionsByTemplateID(input[0].ID, func(revision int, tCr *timestamp.Timestamp) error {
					count = count + 1
					return nil
				})
				if err != nil {
					t.Error(err)
				}
				if count != len(input) {
					t.Errorf("expected %d template revisions stored in the database but we got %d", len(input), count)
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
				_, wName, _, err := tinkDB.GetTemplate(context.Background(), map[string]string{"id": input[0].ID}, false)
				if err != nil {
					t.Error(err)
				}
				if wName != "updated-name" {
					t.Errorf("expected name to be \"%s\", got \"%s\"", "updated-name", wName)
				}

				count := 0
				err = tinkDB.ListRevisionsByTemplateID(input[0].ID, func(revision int, tCr *timestamp.Timestamp) error {
					count = count + 1
					return nil
				})
				if err != nil {
					t.Error(err)
				}
				if count != 2 {
					t.Errorf("expected %d template revisions stored in the database but we got %d", len(input), count)
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
					count = count + 1
					return nil
				})
				if err != nil {
					t.Error(err)
				}
				if len(input) != count {
					t.Errorf("expected %d templates stored in the database but we got %d", len(input), count)
				}

				revisions := 0
				for _, w := range input {
					err := tinkDB.ListRevisionsByTemplateID(w.ID, func(revision int, tCr *timestamp.Timestamp) error {
						revisions = revisions + 1
						return nil
					})
					if err != nil {
						t.Error(err)
					}
				}
				if revisions != len(input) {
					t.Errorf("expected %d template revisions stored in the database but we got %d", len(input), revisions)
				}
			},
		},
	}

	for _, s := range table {
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
	_, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{
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
	_, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{
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

	revisions := 0
	err = tinkDB.ListRevisionsByTemplateID(w.ID, func(revision int, tCr *timestamp.Timestamp) error {
		revisions = revisions + 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if revisions != 1 {
		t.Errorf("expected 1 template revisions stored in the database but we got %d", revisions)
	}

	err = tinkDB.DeleteTemplate(ctx, w.ID)
	if err != nil {
		t.Error(err)
	}

	count := 0
	err = tinkDB.ListTemplates("%", func(id, n string, in, del *timestamp.Timestamp) error {
		count = count + 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if count != 0 {
		t.Errorf("expected 0 templates stored in the database after delete, but we got %d", count)
	}

	revisions = 0
	err = tinkDB.ListRevisionsByTemplateID(w.ID, func(revision int, tCr *timestamp.Timestamp) error {
		revisions = revisions + 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if revisions != 0 {
		t.Errorf("expected 0 template revisions stored in the database but we got %d", revisions)
	}
}

func TestGetTemplate(t *testing.T) {
	ctx := context.Background()
	expectation := func(t *testing.T, input *workflow.Workflow, tinkDB *db.TinkDB) {
		wID, wName, wData, err := tinkDB.GetTemplate(ctx, map[string]string{"id": input.ID}, false)
		if err != nil {
			t.Error(err)
		}
		w := workflow.MustParse([]byte(wData))
		w.ID = wID
		w.Name = wName
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
			_, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{
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
	_, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{
		ApplyMigration: true,
	})
	defer func() {
		err := cl()
		if err != nil {
			t.Error(err)
		}
	}()

	id := uuid.New().String()
	_, _, _, err := tinkDB.GetTemplate(ctx, map[string]string{"id": id}, false)
	if err == nil {
		t.Error("expected error, got nil")
	}

	want := "no rows in result set"
	if !strings.Contains(err.Error(), want) {
		t.Error(fmt.Errorf("unexpected output, looking for %q as a substring in %q", want, err.Error()))
	}
}

func TestUpdateTemplate(t *testing.T) {
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

	// create template
	w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
	w.ID = uuid.New().String()
	w.Name = fmt.Sprintf("id_%d", rand.Int())
	err := createTemplateFromWorkflowType(ctx, tinkDB, w)
	if err != nil {
		t.Error(err)
	}

	// validate revisions
	revisions := 0
	err = tinkDB.ListRevisionsByTemplateID(w.ID, func(revision int, tCr *timestamp.Timestamp) error {
		revisions = revisions + 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if revisions != 1 {
		t.Errorf("expected 1 template revisions stored in the database but we got %d", revisions)
	}

	// update template
	w.Name = "updated-template-name"
	data, err := json.Marshal(w)
	if err != nil {
		t.Error(err)
	}
	err = tinkDB.UpdateTemplate(ctx, w.Name, string(data), uuid.MustParse(w.ID))
	if err != nil {
		t.Error(err)
	}

	// revalidate revisions
	revisions = 0
	err = tinkDB.ListRevisionsByTemplateID(w.ID, func(revision int, tCr *timestamp.Timestamp) error {
		revisions = revisions + 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if revisions != 2 {
		t.Errorf("expected 2 template revisions stored in the database but we got %d", revisions)
	}
}

func TestTemplateRevisionMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{})
	defer func() {
		err := cl()
		if err != nil {
			t.Error(err)
		}
	}()

	// apply old migrations
	_, err := migrate.Exec(db, "postgres", migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			migration.Get202009171251(),
			migration.Get202010071530(),
			migration.Get202010221010(),
			migration.Get202012041103(),
			migration.Get202012091055(),
			migration.Get2020121691335(),
		},
	}, migrate.Up)
	if err != nil {
		t.Error(err)
	}

	// create a few templates
	err = createTemplatesFromSQL(ctx, db)
	if err != nil {
		t.Error(err)
	}

	// check if the templates have been created successfully
	count := 0
	err = tinkDB.ListTemplates("%", func(id, n string, in, del *timestamp.Timestamp) error {
		count = count + 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if count != 5 {
		t.Errorf("expected %d templates stored in the database but we got %d", 5, count)
	}

	// apply template-revision migration
	_, err = migrate.Exec(db, "postgres", migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			migration.Get202009171251(),
			migration.Get202010071530(),
			migration.Get202010221010(),
			migration.Get202012041103(),
			migration.Get202012091055(),
			migration.Get2020121691335(),
			migration.Get202102111035(),
		},
	}, migrate.Up)
	if err != nil {
		t.Error(err)
	}

	// validate if the migration was successful
	revisionCount := 0
	err = tinkDB.ListTemplates("%", func(id, n string, in, del *timestamp.Timestamp) error {
		err := tinkDB.ListRevisionsByTemplateID(id, func(revision int, tCr *timestamp.Timestamp) error {
			revisionCount = revisionCount + 1
			return nil
		})
		return err
	})
	if err != nil {
		t.Error(err)
	}
	if revisionCount != 5 {
		t.Errorf("expected %d template revisions stored in the database but we got %d", 5, revisionCount)
	}
}

// The TestTemplateRevisionMigration fails if we create templates using CreateTemplate
// because it already relies on revision. Even if we don't apply the migration,
// the Go code assumes revision to be in place. In order to fix that you have to add
// a template not via code but with a SQL query.
func createTemplatesFromSQL(ctx context.Context, db *sql.DB) error {
	w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	for i := 1; i <= 5; i++ {
		id := uuid.New().String()
		w.ID = id
		w.Name = fmt.Sprintf("id_%d", rand.Int())
		data, err := json.Marshal(w)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`
		INSERT INTO template (id, name, data, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $4, NULL)
	`, w.ID, w.Name, string(data), time.Now())
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
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
