package cmd

import (
	"fmt"

	"github.com/tinkerbell/tink/cmd/tinkerbell/cmd/hardware"
	"github.com/spf13/cobra"
)

var hardwareCmd = &cobra.Command{
	Use:     "hardware",
	Short:   "tinkerbell hardware client",
	Example: "tinkerbell hardware [command]",
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
