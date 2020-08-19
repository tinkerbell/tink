package main

import (
	"fmt"
	"os"

	"github.com/tinkerbell/tink/cmd/tink-cli/cmd"
)

// version is set at build time
var version = "devel"

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}
