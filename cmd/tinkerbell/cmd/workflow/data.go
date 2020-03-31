package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var (
	version       int32
	needsMetadata bool
	versionOnly   bool
)

// dataCmd represents the data subcommand for workflow command
var dataCmd = &cobra.Command{
	Use:     "data [id]",
	Short:   "get workflow data",
	Example: "tinkerbell workflow data [id] [flags]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires an argument", c.UseLine())
		}
		for _, arg := range args {
			if _, err := uuid.FromString(arg); err != nil {
				return fmt.Errorf("invalid uuid: %s", arg)
			}
		}
		return nil
	},
	Run: func(c *cobra.Command, args []string) {
		for _, arg := range args {
			req := &workflow.GetWorkflowDataRequest{WorkflowID: arg, Version: version}
			var res *workflow.GetWorkflowDataResponse
			var err error
			if needsMetadata {
				res, err = client.WorkflowClient.GetWorkflowMetadata(context.Background(), req)
			} else if versionOnly {
				res, err = client.WorkflowClient.GetWorkflowDataVersion(context.Background(), req)
			} else {
				res, err = client.WorkflowClient.GetWorkflowData(context.Background(), req)
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

func init() {
	flags := dataCmd.PersistentFlags()
	flags.Int32VarP(&version, "version", "v", 0, "data version")
	flags.BoolVarP(&needsMetadata, "metadata", "m", false, "metadata only")
	flags.BoolVarP(&versionOnly, "latest version", "l", false, "latest version")

	SubCommands = append(SubCommands, dataCmd)
}
