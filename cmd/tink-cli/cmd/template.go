package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/template"
)

func NewTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "tink template client",
		Example: "tink template [command]",
		Args: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("%v requires arguments", c.UseLine())
			}
			return nil
		},
	}

	metaClient, err := client.NewMetaClientFromGlobal()
	if err != nil {
		panic(err)
	}

	cmd.AddCommand(template.NewCreateCommand(metaClient))
	cmd.AddCommand(template.NewDeleteCommand(metaClient))
	cmd.AddCommand(template.NewListCommand(metaClient))
	cmd.AddCommand(template.NewUpdateCommand(metaClient))

	// If the variable TINK_CLI_VERSION is not set to 0.0.0 use the old get
	// command. This is a way to keep retro-compatibility with the old get command.
	getCmd := template.GetCmd
	if v := os.Getenv("TINK_CLI_VERSION"); v != "0.0.0" {
		getCmd = template.NewGetTemplateCommand(metaClient)
	}
	cmd.AddCommand(getCmd)
	return cmd
}
