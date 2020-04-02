package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/cli/tink/cmd/hardware"
)

var hardwareCmd = &cobra.Command{
	Use:     "hardware",
	Short:   "tink hardware client",
	Example: "tink hardware [command]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires arguments", c.UseLine())
		}
		return nil
	},
}

func init() {
	hardwareCmd.AddCommand(hardware.SubCommands...)
	rootCmd.AddCommand(hardwareCmd)
}
