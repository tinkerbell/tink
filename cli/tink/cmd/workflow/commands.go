package workflow

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

// SubCommands hold all the subcommands for tinkerbell cli
var SubCommands []*cobra.Command

func validateID(id string) error {
	if _, err := uuid.FromString(id); err != nil {
		return fmt.Errorf("invalid uuid: %s", id)
	}
	return nil
}
