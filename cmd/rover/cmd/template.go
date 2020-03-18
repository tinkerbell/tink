package cmd

import (
	"fmt"

	"github.com/packethost/tinkerbell/cmd/rover/cmd/template"
	"github.com/spf13/cobra"
)

// templateCmd represents the template sub-command
var templateCmd = &cobra.Command{
	Use:     "template",
	Short:   "rover template client",
	Example: "rover template [command]",
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
