package create

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"google.golang.org/grpc"
)

type ResourceType string

var (
	Resources = map[ResourceType]int32{
		"HARDWARE": 0,
		"TEMPLATE": 1,
		"WORKFLOW": 2,
	}
)

// Options struct
type Options struct {

	// Based on this the Create function will be called
	Resource ResourceType
	// CreateByStdin is used for creation of Hardware and Template resources
	CreateByStdin func(context.Context, *client.FullClient, []byte) (interface{}, error)

	// CreateByFlags is used for Workflow resource creation
	CreateByFlags func(context.Context, *client.FullClient) (interface{}, error)

	clientConnOpt *client.ConnOptions
	fullClient    *client.FullClient
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
			if opt.CreateByStdin == nil {
				return errors.New("createByStdin is not implemented for this resource yet. Please have a look at the issue in GitHub or open a new one")
			}
			for _, data := range args {
				resourceID, err := opt.CreateByStdin(cmd.Context(), opt.fullClient, []byte(data))
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Created\t%s\n", resourceID)
			}
			return nil
		},
	}
	if opt.clientConnOpt == nil {
		opt.SetClientConnOpt(&client.ConnOptions{})
	}
	opt.clientConnOpt.SetFlags(cmd.PersistentFlags())
	return cmd
}
