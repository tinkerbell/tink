package get

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"google.golang.org/grpc"
)

type Options struct {
	// Headers is the list of headers you want to print as part of the list
	Headers []string
	// RetrieveData reaches out to Tinkerbell and it gets the required data
	RetrieveData func(context.Context, *client.FullClient) ([]interface{}, error)
	// RetrieveByID is used when a get command has a list of arguments
	RetrieveByID func(context.Context, *client.FullClient, string) (interface{}, error)
	// PopulateTable populates a table with the data retrieved with the RetrieveData function.
	PopulateTable func([]interface{}, table.Writer) error

	clientConnOpt *client.ConnOptions
	fullClient    *client.FullClient

	// Format specifies the format you want the list of resources printed
	// out. By default it is table but it can be JSON ar CSV.
	Format string
	// NoHeaders does not print the header line
	NoHeaders bool
}

func (o *Options) SetClientConnOpt(co *client.ConnOptions) {
	o.clientConnOpt = co
}

func (o *Options) SetFullClient(cl *client.FullClient) {
	o.fullClient = cl
}

const shortDescr = `display one or many resources`

const longDescr = `Prints a table containing the most important information about a specific
resource. You can specify the kind of output you want to receive. It can be
table, csv or json.
`

const exampleDescr = `# List all hardware in table output format.
tink hardware get

# List all workflow in csv output format.
tink template get --format csv

# List a single template in json output format.
tink workflow get --format json [id]
`

func NewGetCommand(opt Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "get",
		Short:                 shortDescr,
		Long:                  longDescr,
		Example:               exampleDescr,
		DisableFlagsInUseLine: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opt.fullClient != nil {
				return nil
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opt.fullClient == nil {
				var err error
				var conn *grpc.ClientConn
				conn, err = client.NewClientConn(opt.clientConnOpt)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Flag based client configuration failed with err: %s. Trying with env var legacy method...", err)
					// Fallback to legacy Setup via env var
					conn, err = client.GetConnection()
					if err != nil {
						return errors.Wrap(err, "failed to setup connection to tink-server")
					}
				}
				opt.SetFullClient(client.NewFullClient(conn))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var data []interface{}

			t := table.NewWriter()
			t.SetOutputMirror(cmd.OutOrStdout())

			if len(args) != 0 {
				if opt.RetrieveByID == nil {
					return errors.New("option RetrieveByID is not implemented for this resource yet. Please have a look at the issue in GitHub or open a new one")
				}
				for _, requestedID := range args {
					s, err := opt.RetrieveByID(cmd.Context(), opt.fullClient, requestedID)
					if err != nil {
						continue
					}
					data = append(data, s)
				}
			} else {
				data, err = opt.RetrieveData(cmd.Context(), opt.fullClient)
			}
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
	cmd.PersistentFlags().BoolVar(&opt.NoHeaders, "no-headers", false, "Table contains an header with the columns' name. You can disable it from being printed out")
	if opt.clientConnOpt == nil {
		opt.SetClientConnOpt(&client.ConnOptions{})
	}
	opt.clientConnOpt.SetFlags(cmd.PersistentFlags())
	return cmd
}
