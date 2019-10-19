package workflow

import (
	"fmt"

	_ "github.com/packethost/rover/client"
	"github.com/spf13/cobra"
)

var (
	fTemplate = "template"
	fTarget   = "target"
	template  string
	target    string
)

// createCmd represents the create sub command
var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a workflow",
	Example: "workflow create [flags]",
	Run: func(c *cobra.Command, args []string) {
		validateUUID()
		createWorkflow(c, args)
		fmt.Println("executing create workflow")
	},
}

func addFlags() {
	flags := createCmd.PersistentFlags()
	flags.StringVarP(&template, "template", "t", "", "workflow template")
	flags.StringVarP(&target, "target", "r", "", "workflow target")

	createCmd.MarkPersistentFlagRequired(fTarget)
	createCmd.MarkPersistentFlagRequired(fTemplate)
}

// TODO
func validateUUID() {}

func createWorkflow(c *cobra.Command, args []string) {
	// _ := client.ConnectGRPC(c.Flags().GetString("facility"))

}

func init() {
	addFlags()
	SubCommands = append(SubCommands, createCmd)
}
