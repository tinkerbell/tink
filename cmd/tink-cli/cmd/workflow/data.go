package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
)

var (
	version       int32
	needsMetadata bool
	versionOnly   bool
)

// dataCmd represents the data subcommand for workflow command
func NewDataCommand(cl *client.MetaClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "data [id]",
		Short:   "get workflow data",
		Example: "tink workflow data [id] [flags]",
		Args: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("%v requires an argument", c.UseLine())
			}
			for _, arg := range args {
				if _, err := uuid.Parse(arg); err != nil {
					return fmt.Errorf("invalid uuid: %s", arg)
				}
			}
			return nil
		},
		Run: func(c *cobra.Command, args []string) {
			for _, arg := range args {
				req := &workflow.GetWorkflowDataRequest{WorkflowId: arg, Version: version}
				var res *workflow.GetWorkflowDataResponse
				var err error
				if needsMetadata {
					res, err = cl.WorkflowClient.GetWorkflowMetadata(context.Background(), req)
				} else if versionOnly {
					res, err = cl.WorkflowClient.GetWorkflowDataVersion(context.Background(), req)
				} else {
					res, err = cl.WorkflowClient.GetWorkflowData(context.Background(), req)
				}

				if err != nil {
					log.Fatal(err)
				}

				if versionOnly {
					fmt.Printf("Latest workflow data version: %v\n", res.Version)
				} else {
					fmt.Println(string(res.Data))
				}
			}
		},
	}
	flags := cmd.PersistentFlags()
	flags.Int32VarP(&version, "version", "v", 0, "data version")
	flags.BoolVarP(&needsMetadata, "metadata", "m", false, "metadata only")
	flags.BoolVarP(&versionOnly, "latest version", "l", false, "latest version")

	return cmd
}
