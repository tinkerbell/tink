package cmd

import (
	"fmt"

	"github.com/tinkerbell/tink/cmd/tinkerbell/cmd/template"
	"github.com/spf13/cobra"
)

// templateCmd represents the template sub-command
var templateCmd = &cobra.Command{
	Use:     "template",
	Short:   "tinkerbell template client",
	Example: "tinkerbell template [command]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires arguments", c.UseLine())
		}
		return nil
	},
}

func init() {
	templateCmd.AddCommand(template.SubCommands...)
	rootCmd.AddCommand(templateCmd)
}
