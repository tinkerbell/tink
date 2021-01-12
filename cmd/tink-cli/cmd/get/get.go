package get

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

type CmdOpt struct {
	// Headers is the list of headers you want to print as part of the list
	Headers []string
	// RetrieveData reaches out to Tinkerbell and it gets the required data
	RetrieveData func(context.Context) ([]interface{}, error)
	// PopulateTable populates a table with the data retrieved with the RetrieveData function.
	PopulateTable func([]interface{}, table.Writer) error

	// Format specifies the format you want the list of resources printed
	// out. By default it is table but it can be JSON ar CSV.
	Format string
	// NoHeaders does not print the header line
	NoHeaders bool
}

const shortDescr = `
Display one or many resources`

const longDescr = `
Prints a table contains the most important information about a specific
resource. You can specify the kind of output you want to receive. It can be
table or csv or json.
`

const exampleDescr = `
# List all hardware in table output format.
tink hardware get

# List all workflow in csv output format.
tink template get --format csv

# List a single template in json output format.
tink workflow get --format json [id]
`

func NewGetCommand(opt CmdOpt) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "get",
		Short:                 shortDescr,
		Long:                  longDescr,
		Example:               exampleDescr,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			t := table.NewWriter()
			t.SetOutputMirror(cmd.OutOrStdout())

			data, err := opt.RetrieveData(cmd.Context())
			if err != nil {
				return err
			}

			if !opt.NoHeaders {
				header := table.Row{}
				for _, h := range opt.Headers {
					header = append(header, h)
				}
				t.AppendHeader(header)
			}

			// TODO(gianarb): Technically this is not needed for
			// all the output formats but for now that's fine
			if err := opt.PopulateTable(data, t); err != nil {
				return err
			}

			switch opt.Format {
			case "json":
				// TODO(gianarb): the table library we use do
				// not support JSON right now. I am not even
				// sure I like tables! So complicated...
				b, err := json.Marshal(struct {
					Data interface{} `json:"data"`
				}{Data: data})
				if err != nil {
					return err
				}
				fmt.Fprint(cmd.OutOrStdout(), string(b))
			case "csv":
				t.RenderCSV()
			default:
				t.Render()
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&opt.Format, "format", "", "table", "The format you expect the list to be printed out. Currently supported format are table, JSON and CSV")
	cmd.PersistentFlags().BoolVar(&opt.NoHeaders, "no-headers", false, "Table contains an header with the columns' name. You can disable it from behing printed out")
	return cmd
}
