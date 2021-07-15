package main

import (
	"os"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/cmd/tink-worker/cmd"
	"github.com/tobert/otel-launcher-go/launcher"
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

	otel := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("github.com/tinkerbell/tink"),
	)
	defer otel.Shutdown()

	rootCmd := cmd.NewRootCommand(version, logger)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
