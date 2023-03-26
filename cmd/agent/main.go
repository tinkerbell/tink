package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tinkerbell/tink/internal/cli"
)

func main() {
	agent := cli.NewAgent()
	if err := agent.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
