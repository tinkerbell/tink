package integration

import (
	"context"
	"testing"

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
		// Expectation is the function used to apply the assertions.
		// You can use it to validate if the Input are created as you expect
		Expectation func(*testing.T, []*workflow.Workflow, *db.TinkDB)
	}{
		{
			Name: "happy-path-single-crete-template",
			Input: []*workflow.Workflow{
				func() *workflow.Workflow {
					w := workflow.MustParseFromFile("./testdata/template_happy_path_1.yaml")
					w.ID = "545f7ce9-5313-49c6-8704-0ed98814f1f7"
					return w
				}(),
			},
			Expectation: func(t *testing.T, input []*workflow.Workflow, tinkDB *db.TinkDB) {
				wID, wName, wData, err := tinkDB.GetTemplate(context.Background(), map[string]string{"id": input[0].ID}, false)
				if err != nil {
					t.Error(err)
				}
				w := workflow.MustParse([]byte(wData))
				w.ID = wID
				w.Name = wName
				if dif := cmp.Diff(input[0], w); dif != "" {
					t.Errorf(dif)
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
			defer cl()
			for _, tt := range s.Input {
				uID := uuid.MustParse(tt.ID)
				content, err := yaml.Marshal(tt)
				if err != nil {
					t.Error(err)
				}
				err = tinkDB.CreateTemplate(ctx, tt.Name, string(content), uID)
				if err != nil {
					t.Error(err)
				}
			}
			s.Expectation(t, s.Input, tinkDB)
		})

	}
}
