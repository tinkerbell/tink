package main

import (
	"fmt"
	"os"

	"github.com/tinkerbell/tink/cmd/tink-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}
