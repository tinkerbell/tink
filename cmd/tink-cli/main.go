package main

import (
	"context"
	"fmt"
	"os"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd"
)

// version is set at build time.
var version = "devel"

func main() {
	ctx, otelShutdown := otelinit.InitOpenTelemetry(context.Background(), "github.com/tinkerbell/tink")

	if err := cmd.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	otelShutdown(ctx)
}
