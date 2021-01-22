package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
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

	cl, err := client.NewFullClientFromGlobal()
	if err != nil {
		panic(err)
	}

	cmd.AddCommand(template.NewCreateCommand(cl))
	cmd.AddCommand(template.NewDeleteCommand(cl))
	cmd.AddCommand(template.NewListCommand(cl))
	cmd.AddCommand(template.NewUpdateCommand(cl))

	// If the variable TINK_CLI_VERSION is not set to 0.0.0 use the old get
	// command. This is a way to keep retro-compatibility with the old get command.
	getCmd := template.GetCmd
	if v := os.Getenv("TINK_CLI_VERSION"); v != "0.0.0" {
		getCmd = get.NewGetCommand(template.NewGetOptions())
	}
	cmd.AddCommand(getCmd)
	return cmd
}
