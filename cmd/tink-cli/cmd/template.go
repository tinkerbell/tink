package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/template"
)

// templateCmd represents the template sub-command
var templateCmd = &cobra.Command{
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

func init() {
	templateCmd.AddCommand(template.SubCommands...)
	rootCmd.AddCommand(templateCmd)
}
