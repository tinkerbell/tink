package create

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tinkerbell/tink/client"
	"google.golang.org/grpc"
)

var (
	ErrStdinCollisionWithStdinAndFilePath = fmt.Errorf("It looks like you are passing input even as stdin and as file. You have to choose one of them.")
)

type Options struct {
	FilePath string
	FlagSet  *pflag.FlagSet
	// MergeValidateAndCreateFunc contains the logic coming from the
	// actual resource itself. It gets a reader with the content
	// coming from stdin or from the specified file.
	// If the reader is empty it means that all the values are coming
	// from flags. It is reponsability for this function to parse the
	// input, validate it, merge them accordigly and save them.
	// It returns an error to notify that something didn't go as expected for example:
	// * The reader is empty and there are no flags, so there is
	// nothing to do, input is not valid.
	// * The resuled object is invalid
	// * Error from the grpc-server
	// It returns the ID of the created resource and it will get
	// printed to stdout.
	MergeValidateAndCreateFunc func(io.Reader) (id string, err error)
	clientConnOpt              *client.ConnOptions
	fullClient                 *client.FullClient
}

// SetClientConnOpt set client options
func (o *Options) SetClientConnOpt(co *client.ConnOptions) {
	o.clientConnOpt = co
}

// SetFullClient set complete clim
func (o *Options) SetFullClient(cl *client.FullClient) {
	o.fullClient = cl
}

const shortDescr = "Create one or more resources"

const longDescr = `Creates one or more resources and prints the status of
the deleted resource.

# Create template resource (success)
tink template create < <path to template file>
Created Template: 8ae1cc24-6a9c-11eb-a0fc-0242ac120005

# Create template resource (not found)
tink template create < <path to template file>
Error	file not found or
`

const exampleDescr = `# Create template resource
tink template create < <path to template file>

# Create hardware resource
tink hardware create < <path to hardware data file>

# Create workflow resource
tink workflow create --template-id <template id> --hardware '{"device_1":"00:99:88:77:66"}'

`

// NewCreateCommand creates a new `tink create` command
func NewCreateCommand(opt Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "create",
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
			stdin, err := ioutil.ReadAll(cmd.InOrStdin())
			if err != nil {
				return nil
			}

			if len(stdin) != 0 && opt.FilePath != "" {
				return ErrStdinCollisionWithStdinAndFilePath
			}
			var in io.Reader
			in = bytes.NewReader(stdin)
			if opt.FilePath != "" {
				f, err := os.Open(opt.FilePath)
				if err != nil {
					return err
				}
				in = f
			}
			id, err := opt.MergeValidateAndCreateFunc(in)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&opt.FilePath, "file", "", "Location of the file used as input.")
	cmd.PersistentFlags().AddFlagSet(opt.FlagSet)
	if opt.clientConnOpt == nil {
		opt.SetClientConnOpt(&client.ConnOptions{})
	}
	opt.clientConnOpt.SetFlags(cmd.PersistentFlags())
	return cmd
}
