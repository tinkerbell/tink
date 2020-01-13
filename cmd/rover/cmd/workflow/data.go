package workflow

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/workflow"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var (
	version  string
	fVersion = "version"
)

// dataCmd represents the data subcommand for workflow command
var dataCmd = &cobra.Command{
	Use:     "data [id]",
	Short:   "get workflow data",
	Example: "rover workflow data [id] [flags]",
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
			req := &workflow.GetWorkflowDataRequest{WorkflowID: arg}
			if version != "" {
				v, err := strconv.ParseInt(version, 10, 64)
				if err != nil {
					log.Fatal(fmt.Errorf("invalid version: %v", version))
					return
				}
				req.Version = v
			}
			res, err := client.WorkflowClient.GetWorkflowData(context.Background(), req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(res.Data))
		}
	},
}

func init() {
	flags := dataCmd.PersistentFlags()
	flags.StringVarP(&version, fVersion, "v", "", "data version")

	SubCommands = append(SubCommands, dataCmd)
}
