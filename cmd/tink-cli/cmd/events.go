package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/events"
)

var eventcmd = &cobra.Command{
	Use:     "events",
	Short:   "tink events client",
	Example: "tink events [command]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires arguments", c.UseLine())
		}
		return nil
	},
}

func init() {
	eventcmd.AddCommand(events.SubCommands...)
	rootCmd.AddCommand(eventcmd)
}
