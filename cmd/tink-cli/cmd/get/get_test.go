package get

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jedib0t/go-pretty/table"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/internal/clientctx"
)

func TestNewGetCommand(t *testing.T) {
	tests := []struct {
		Name         string
		ExpectStdout string
		ExpectError  error
		Args         []string
		Opt          Options
		Skip         string
		Run          func(t *testing.T, cmd *cobra.Command, stdout, stderr io.Reader)
	}{
		{
			Name: "happy-path",
			ExpectStdout: `+------+-------+
| NAME | ID    |
+------+-------+
| 10   | hello |
+------+-------+
`,
			Opt: Options{
				Headers: []string{"name", "id"},
				RetrieveData: func(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
					data := []interface{}{
						[]string{"10", "hello"},
					}
					return data, nil
				},
				PopulateTable: func(data []interface{}, w table.Writer) error {
					for _, v := range data {
						if vv, ok := v.([]string); ok {
							w.AppendRow(table.Row{vv[0], vv[1]})
						}
					}
					return nil
				},
			},
		},
		{
			Name: "get-by-id",
			Args: []string{"e0ffbf50-ae7c-4c92-bc7f-34e0de25a989"},
			ExpectStdout: `+-------+--------------------------------------+
| NAME  | ID                                   |
+-------+--------------------------------------+
| hello | e0ffbf50-ae7c-4c92-bc7f-34e0de25a989 |
+-------+--------------------------------------+
`,
			Opt: Options{
				Headers: []string{"name", "id"},
				RetrieveByID: func(ctx context.Context, cl *client.FullClient, arg string) (interface{}, error) {
					if arg != "e0ffbf50-ae7c-4c92-bc7f-34e0de25a989" {
						t.Errorf("expected e0ffbf50-ae7c-4c92-bc7f-34e0de25a989 as arg got %s", arg)
					}
					return []string{"hello", "e0ffbf50-ae7c-4c92-bc7f-34e0de25a989"}, nil
				},
				PopulateTable: func(data []interface{}, w table.Writer) error {
					for _, v := range data {
						if vv, ok := v.([]string); ok {
							w.AppendRow(table.Row{vv[0], vv[1]})
						}
					}
					return nil
				},
			},
		},
		{
			Name:        "get-by-id but no retriever",
			Args:        []string{"e0ffbf50-ae7c-4c92-bc7f-34e0de25a989"},
			ExpectError: errors.New("get by ID is not implemented for this resource yet, please have a look at the issue in GitHub or open a new one"),
		},
		{
			Name: "get-by-name",
			Args: []string{"hello"},
			ExpectStdout: `+-------+--------------------------------------+
| NAME  | ID                                   |
+-------+--------------------------------------+
| hello | e0ffbf50-ae7c-4c92-bc7f-34e0de25a989 |
+-------+--------------------------------------+
`,
			Opt: Options{
				Headers: []string{"name", "id"},
				RetrieveByName: func(ctx context.Context, cl *client.FullClient, arg string) (interface{}, error) {
					if arg != "hello" {
						t.Errorf("expected hello as arg got %s", arg)
					}
					return []string{"hello", "e0ffbf50-ae7c-4c92-bc7f-34e0de25a989"}, nil
				},
				PopulateTable: func(data []interface{}, w table.Writer) error {
					for _, v := range data {
						if vv, ok := v.([]string); ok {
							w.AppendRow(table.Row{vv[0], vv[1]})
						}
					}
					return nil
				},
			},
		},
		{
			Name:        "get-by-name but no retriever",
			Args:        []string{"hello"},
			ExpectError: errors.New("get by Name is not implemented for this resource yet, please have a look at the issue in GitHub or open a new one"),
		},
		{
			Name: "happy-path-no-headers",
			ExpectStdout: `+----+-------+
| 10 | hello |
+----+-------+
`,
			Args: []string{"--no-headers"},
			Opt: Options{
				Headers: []string{"name", "id"},
				RetrieveData: func(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
					data := []interface{}{
						[]string{"10", "hello"},
					}
					return data, nil
				},
				PopulateTable: func(data []interface{}, w table.Writer) error {
					for _, v := range data {
						if vv, ok := v.([]string); ok {
							w.AppendRow(table.Row{vv[0], vv[1]})
						}
					}
					return nil
				},
			},
		},
		{
			Name:         "happy-path-json",
			ExpectStdout: `{"data":[["10","hello"]]}`,
			Args:         []string{"--format", "json"},
			Opt: Options{
				Headers: []string{"name", "id"},
				RetrieveData: func(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
					data := []interface{}{
						[]string{"10", "hello"},
					}
					return data, nil
				},
				PopulateTable: func(data []interface{}, w table.Writer) error {
					for _, v := range data {
						if vv, ok := v.([]string); ok {
							w.AppendRow(table.Row{vv[0], vv[1]})
						}
					}
					return nil
				},
			},
		},
		{
			Name: "happy-path-json-no-headers",
			Skip: "The JSON format is rusty and custom because the table library we use does not support JSON right now. This feature is not implemented.",
		},
		{
			Name: "happy-path-csv-no-headers",
			ExpectStdout: `10,hello
`,
			Args: []string{"--format", "csv", "--no-headers"},
			Opt: Options{
				Headers: []string{"name", "id"},
				RetrieveData: func(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
					data := []interface{}{
						[]string{"10", "hello"},
					}
					return data, nil
				},
				PopulateTable: func(data []interface{}, w table.Writer) error {
					for _, v := range data {
						if vv, ok := v.([]string); ok {
							w.AppendRow(table.Row{vv[0], vv[1]})
						}
					}
					return nil
				},
			},
		},
		{
			Name: "happy-path-csv",
			ExpectStdout: `name,id
10,hello
`,
			Args: []string{"--format", "csv"},
			Opt: Options{
				Headers: []string{"name", "id"},
				RetrieveData: func(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
					data := []interface{}{
						[]string{"10", "hello"},
					}
					return data, nil
				},
				PopulateTable: func(data []interface{}, w table.Writer) error {
					for _, v := range data {
						if vv, ok := v.([]string); ok {
							w.AppendRow(table.Row{vv[0], vv[1]})
						}
					}
					return nil
				},
			},
		},
		{
			Name:        "no opts",
			ExpectError: errors.New("get-all-data is not implemented for this resource yet, please have a look at the issue in GitHub or open a new one"),
		},
	}

	for _, s := range tests {
		t.Run(s.Name, func(t *testing.T) {
			if s.Skip != "" {
				t.Skip(s.Skip)
			}
			stdout := &bytes.Buffer{}
			cmd := NewGetCommand(s.Opt)
			cmd.SilenceErrors = true
			cmd.SetOut(stdout)
			cmd.SetArgs(s.Args)
			err := cmd.ExecuteContext(clientctx.Set(context.Background(), &client.FullClient{}))
			if fmt.Sprint(err) != fmt.Sprint(s.ExpectError) {
				t.Errorf("unexpected error: want=%v, got=%v", s.ExpectError, err)
			}
			out, err := ioutil.ReadAll(stdout)
			if err != nil {
				t.Error(err)
			}
			if s.ExpectError != nil {
				s.ExpectStdout = cmd.UsageString() + "\n"
			}
			if diff := cmp.Diff(string(out), s.ExpectStdout); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
