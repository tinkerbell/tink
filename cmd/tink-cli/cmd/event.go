package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/event"
)

var eventcmd = &cobra.Command{
	Use:     "event",
	Short:   "tink event client",
	Example: "tink event [command]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires arguments", c.UseLine())
		}
		return nil
	},
}

func init() {
	eventcmd.AddCommand(event.SubCommands...)
	rootCmd.AddCommand(eventcmd)
}
