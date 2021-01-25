package workflow

import (
	"fmt"

	"github.com/google/uuid"
)

func validateID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid uuid: %s", id)
	}
	return nil
}
