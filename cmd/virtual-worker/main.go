package main

import (
	"os"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/cmd/virtual-worker/cmd"
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
	rootCmd := cmd.NewRootCommand(version, logger)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	logger.Close()
}
