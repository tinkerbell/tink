package main

import (
	"os"

	"github.com/tinkerbell/tink/internal/cli"
)

func main() {
	if err := cli.NewAgent().Execute(); err != nil {
		os.Exit(-1)
	}
}
