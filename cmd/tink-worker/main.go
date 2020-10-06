package main

import (
	"os"

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

	rootCmd := cmd.NewRootCommand(version, logger)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
