package main

import (
	"context"
	"os"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/cmd/tink-worker/cmd"
)

const (
	serviceKey = "github.com/tinkerbell/tink"
)

// version is set at build time.
var version = "devel"

func main() {
	logger, err := log.Init(serviceKey)
	if err != nil {
		panic(err)
	}

	ctx, otelShutdown := otelinit.InitOpenTelemetry(context.Background(), "github.com/tinkerbell/tink")

	rootCmd := cmd.NewRootCommand(version, logger)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	logger.Close()
	otelShutdown(ctx)
}
