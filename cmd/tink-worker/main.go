package main

import (
	"fmt"
	"os"

	"github.com/tinkerbell/tink/cmd/tink-worker/cmd"
)

func main() {
	// parse and validate command-line flags and required env vars
	flagEnvSettings, err := cmd.CollectFlagEnvSettings(os.Args[1:])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Remove me: this is just here during developent to prevent linter issues
	fmt.Printf("Version flag is %v\n", flagEnvSettings.Version)
}
