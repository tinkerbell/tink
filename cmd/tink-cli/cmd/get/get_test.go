package get

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

func TestNewGetCommand(t *testing.T) {
	table := []struct {
		Name         string
		ExpectStdout string
		Args         []string
		Opt          CmdOpt
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
			Opt: CmdOpt{
				Headers: []string{"name", "id"},
				RetrieveData: func(ctx context.Context) ([]interface{}, error) {
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
			Name: "happy-path-no-headers",
			ExpectStdout: `+----+-------+
| 10 | hello |
+----+-------+
`,
			Args: []string{"--no-headers"},
			Opt: CmdOpt{
				Headers: []string{"name", "id"},
				RetrieveData: func(ctx context.Context) ([]interface{}, error) {
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
			Opt: CmdOpt{
				Headers: []string{"name", "id"},
				RetrieveData: func(ctx context.Context) ([]interface{}, error) {
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
	}

	for _, s := range table {
		t.Run(s.Name, func(t *testing.T) {
			stdout := bytes.NewBufferString("")
			cmd := NewGetCommand(s.Opt)
			cmd.SetOut(stdout)
			cmd.SetArgs(s.Args)
			err := cmd.Execute()
			if err != nil {
				t.Error(err)
			}
			out, err := ioutil.ReadAll(stdout)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(string(out), s.ExpectStdout); diff != "" {
				t.Fatal(diff)
			}
		})
	}

}
