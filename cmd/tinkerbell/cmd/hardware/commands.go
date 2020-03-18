package hardware

import (
	"fmt"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

// SubCommands holds the sub commands for template command
// Example: tinkerbell template [subcommand]
var SubCommands []*cobra.Command

func verifyUUIDs(args []string) error {
	if len(args) < 1 {
		return errors.New("requires at least one id")
	}
	for _, arg := range args {
		if _, err := uuid.FromString(arg); err != nil {
			return fmt.Errorf("invalid uuid: %s", arg)
		}
	}
	return nil
}
