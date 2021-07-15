package main

import (
	"fmt"
	"os"

	"github.com/tinkerbell/tink/cmd/tink-cli/cmd"
	"github.com/tobert/otel-launcher-go/launcher"
)

// version is set at build time
var version = "devel"

func main() {
	otel := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("github.com/tinkerbell/tink"),
	)
	defer otel.Shutdown()

	if err := cmd.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
