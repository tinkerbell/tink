package delete // nolint:predeclared // package name delete has same name as predeclared identifier

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/internal/clientctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Options struct {
	// DeleteByID is used to delete a resource
	DeleteByID func(context.Context, *client.FullClient, string) (interface{}, error)
}

const shortDescr = "delete one or more resources"

const longDescr = `Deletes one or more resources and prints the status of
the deleted resource.

# Delete template resource (success)
tink template delete 8ae1cc24-6a9c-11eb-a0fc-0242ac120005
Deleted	8ae1cc24-6a9c-11eb-a0fc-0242ac120005

# Delete template resource (not found)
tink template delete 8ae1cc24-6a9c-11eb-a0fc-0242ac120005
Error	8ae1cc24-6a9c-11eb-a0fc-0242ac120005	not found

# Delete template resources (one not found)
tink template delete 8ae1cc24-6a9c-11eb-a0fc-0242ac120005 e4115856-4358-429d-a8f6-9e1b7d794b72
Deleted	8ae1cc24-6a9c-11eb-a0fc-0242ac120005
Error	e4115856-4358-429d-a8f6-9e1b7d794b72	not found

# Delete resources and extract resource ID with awk
tink template delete 8ae1cc24-6a9c-11eb-a0fc-0242ac120005 e4115856-4358-429d-a8f6-9e1b7d794b72 | awk {print $2} > result
cat result
8ae1cc24-6a9c-11eb-a0fc-0242ac120005
e4115856-4358-429d-a8f6-9e1b7d794b72
`

const exampleDescr = `# Delete template resource
tink template delete [id]

# Delete hardware resource
tink hardware delete [id]

# Delete workflow resource
tink workflow delete [id]

# Delete multiple workflow resources
tink workflow delete [id_1] [id_2] [id_3]
`

func NewDeleteCommand(opt Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "delete",
		Short:                 shortDescr,
		Long:                  longDescr,
		Example:               exampleDescr,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opt.DeleteByID == nil {
				return errors.New("option DeleteByID is not implemented for this resource yet. Please have a look at the issue in GitHub or open a new one")
			}

			client := clientctx.Get(cmd.Context())
			for _, requestedID := range args {
				_, err := opt.DeleteByID(cmd.Context(), client, requestedID)
				if err != nil {
					if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
						fmt.Fprintf(cmd.ErrOrStderr(), "Error\t%s\tnot found\n", requestedID)
						continue
					}
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Deleted\t%s\n", requestedID)
			}
			return nil
		},
	}
	return cmd
}
