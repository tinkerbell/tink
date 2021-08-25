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

var (
	// version is set at build time
	version = "devel"
)

func main() {
	logger, err := log.Init(serviceKey)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	ctx, otelShutdown := otelinit.InitOpenTelemetry(context.Background(), "github.com/tinkerbell/tink")
	defer otelShutdown(ctx)

	rootCmd := cmd.NewRootCommand(version, logger)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
