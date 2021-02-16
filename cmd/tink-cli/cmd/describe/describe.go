package describe

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
)

type Options struct {
	// DescribeByID is used to describe a resource
	DescribeByID func(context.Context, *client.FullClient, string) (interface{}, error)

	clientConnOpt *client.ConnOptions
	fullClient    *client.FullClient

	// Format specifies the format you want the list of resources printed
	// out.
	Format string
	// ShowEvents prints out events related to the resource
	ShowEvents bool
}

func (o *Options) SetClientConnOpt(co *client.ConnOptions) {
	o.clientConnOpt = co
}

func (o *Options) SetFullClient(cl *client.FullClient) {
	o.fullClient = cl
}

const shortDescr = "describe one resource"

const longDescr = `Describe one resource and prints a detailed description of the selected
resource, including related events.

# Describe template resource
tink template describe --show-events 8ae1cc24-6a9c-11eb-a0fc-0242ac120005

ID:             8ae1cc24-6a9c-11eb-a0fc-0242ac120005
Template:       46b4948d-c762-4f99-9152-49a70c0732e1
Hardware:       b97263cc-b9be-4cdf-92c1-5030fce5d7ea
State:          Running
Creation Time:  Thu, 26 Nov 2020 13:25:30 +0100
Updated  Time:  Thu, 26 Nov 2020 19:25:30 +0100
Events:
  Type       Age
  ----       ----
  Created    1h
`

const exampleDescr = `# Describe template resource
tink template describe [id]

# Describe hardware resource
tink hardware describe [id]

# Describe workflow resource
tink workflow describe [id]

# Describe a workflow and display events related to it.
tink workflow describe --show-events [id]
`

func NewDescribeCommand(opt Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "describe",
		Short:                 shortDescr,
		Long:                  longDescr,
		Example:               exampleDescr,
		DisableFlagsInUseLine: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opt.fullClient != nil {
				return nil
			}
			if opt.clientConnOpt == nil {
				opt.SetClientConnOpt(&client.ConnOptions{})
			}
			opt.clientConnOpt.SetFlags(cmd.PersistentFlags())
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opt.fullClient == nil {
				var err error
				var conn *grpc.ClientConn
				conn, err = client.NewClientConn(opt.clientConnOpt)
				if err != nil {
					println("Flag based client configuration failed with err: %s. Trying with env var legacy method...", err)
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
			if opt.DescribeByID == nil {
				return errors.New("DescribeByID is not implemented for this resource yet. Please have a look at the issue in GitHub or open a new one.")
			}

			if len(args) != 0 {
				requestedID := args[0]
				_, err := opt.DescribeByID(cmd.Context(), opt.fullClient, requestedID)

				if err != nil {
					if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
						fmt.Fprintf(cmd.ErrOrStderr(), "Error\t%s\tnot found\n", requestedID)
					} else {
						return err
					}
				}
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&opt.Format, "format", "", "", "The format you expect the list to be printed out. Currently supported format are human, JSON")
	cmd.PersistentFlags().BoolVar(&opt.ShowEvents, "show-events", false, "If true, display events related to the described object.")
	return cmd
}
