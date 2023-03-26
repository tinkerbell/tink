package cli_test

import (
	"context"
	"testing"

	"github.com/tinkerbell/tink/internal/cli"
)

func TestAgent(t *testing.T) {
	agent := cli.NewAgent()
	if err := agent.ParseAndRun(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}
