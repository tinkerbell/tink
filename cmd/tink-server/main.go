package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/packethost/pkg/log"
	rpcServer "github.com/tinkerbell/tink/grpc-server"
	httpServer "github.com/tinkerbell/tink/http-server"
)

var (
	// version is set at build time
	version = "devel"

	logger log.Logger
)

func main() {
	log, cleanup, err := log.Init("github.com/tinkerbell/tink")

	if err != nil {
		panic(err)
	}
	logger = log
	defer cleanup()

	log.Info("starting version " + version)

	ctx, closer := context.WithCancel(context.Background())
	errCh := make(chan error, 2)
	facility := os.Getenv("FACILITY")

	cert, modT := rpcServer.SetupGRPC(ctx, logger, facility, errCh)
	httpServer.SetupHTTP(ctx, logger, cert, modT, errCh)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	select {
	case err = <-errCh:
		logger.Error(err)
		panic(err)
	case sig := <-sigs:
		logger.With("signal", sig.String()).Info("signal received, stopping servers")
	}
	closer()

	// wait for grpc server to shutdown
	err = <-errCh
	if err != nil {
		logger.Error(err)
		panic(err)
	}
	err = <-errCh
	if err != nil {
		logger.Error(err)
		panic(err)
	}
}
