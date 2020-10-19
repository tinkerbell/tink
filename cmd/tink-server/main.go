package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/db"
	rpcServer "github.com/tinkerbell/tink/grpc-server"
	httpServer "github.com/tinkerbell/tink/http-server"
)

var (
	// version is set at build time
	version = "devel"

	logger log.Logger
)

func main() {
	log, err := log.Init("github.com/tinkerbell/tink")
	if err != nil {
		panic(err)
	}
	logger = log
	defer logger.Close()
	log.Info("starting version " + version)

	ctx, closer := context.WithCancel(context.Background())
	errCh := make(chan error, 2)
	facility := os.Getenv("FACILITY")

	tinkDB := db.Connect(logger)

	_, onlyMigration := os.LookupEnv("ONLY_MIGRATION")
	if onlyMigration {
		logger.Info("Applying migrations. This process will end when migrations will take place.")
		numAppliedMigrations, err := tinkDB.Migrate()
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
		log.With("num_applied_migrations", numAppliedMigrations).Info("Migrations applied successfully")
		os.Exit(0)
	}

	go tinkDB.PurgeEvents(errCh)

	numAvailableMigrations, err := tinkDB.CheckRequiredMigrations()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	if numAvailableMigrations != 0 {
		log.Info("Your database schema is not up to date. Please apply migrations running tink-server with env var ONLY_MIGRATION set.")
	}

	cert, modT := rpcServer.SetupGRPC(ctx, logger, facility, tinkDB, errCh)
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
		log.Fatal(err)
		panic(err)
	}
	err = <-errCh
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
