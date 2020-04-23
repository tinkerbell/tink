package main

import (
	"fmt"
	"os"

	"github.com/tinkerbell/tink/cmd/tink-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("%s", err.Error()))
		os.Exit(1)
	}
}
