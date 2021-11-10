package main

import (
	"fmt"
	"os"

	"github.com/tinkerbell/tink/cmd/tink-worker/cmd"
)

func main() {
	cmdlineFlags, err := cmd.CollectCmdlineFlags(os.Args[1:])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Remove me: this is just here to prevent linter issues
	fmt.Printf("Version flag is %v\n", cmdlineFlags.Version)
}
