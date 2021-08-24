package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/delete"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/hardware"
)

func NewHardwareCommand() *cobra.Command {
	cmd := &cobra.Command{
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

	cmd.AddCommand(get.NewGetCommand(hardware.NewGetOptions()))
	cmd.AddCommand(delete.NewDeleteCommand(hardware.NewDeleteOptions()))
	cmd.AddCommand(hardware.NewGetByIDCmd())
	cmd.AddCommand(hardware.NewGetByIPCmd())
	cmd.AddCommand(hardware.NewListCmd())
	cmd.AddCommand(hardware.NewGetByMACCmd())
	cmd.AddCommand(hardware.NewPushCmd())

	return cmd
}
