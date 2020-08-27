package main

import (
	"os"
	"strconv"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/client"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
)

const (
	retryIntervalDefault = 3
	retryCountDefault    = 3

	serviceKey           = "github.com/tinkerbell/tink"
	invalidRetryInterval = "invalid RETRY_INTERVAL, using default (seconds)"
	invalidMaxRetry      = "invalid MAX_RETRY, using default"

	errWorker = "worker finished with error"
)

var (
	rClient       pb.WorkflowSvcClient
	retryInterval time.Duration
	retries       int
	logger        log.Logger

	// version is set at build time
	version = "devel"
)

func main() {
	log, cleanup, err := log.Init(serviceKey)
	if err != nil {
		panic(err)
	}
	logger = log
	defer logger.Close()
	log.With("version", version).Info("starting")
	setupRetry()
	if setupErr := client.Setup(); setupErr != nil {
		log.Error(setupErr)
		os.Exit(1)
	}
	conn, err := tryClientConnection()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	rClient = pb.NewWorkflowSvcClient(conn)
	err = processWorkflowActions(rClient)
	if err != nil {
		log.Error(errors.Wrap(err, errWorker))
	}
}

func tryClientConnection() (*grpc.ClientConn, error) {
	var err error
	for r := 1; r <= retries; r++ {
		c, e := client.GetConnection()
		if e != nil {
			err = e
			logger.With("error", err, "duration", retryInterval).Info("failed to connect, sleeping before retrying")
			<-time.After(retryInterval * time.Second)
			continue
		}
		return c, nil
	}
	return nil, err
}

func setupRetry() {
	interval := os.Getenv("RETRY_INTERVAL")
	if interval == "" {
		logger.With("default", retryIntervalDefault).Info("RETRY_INTERVAL not set")
		retryInterval = retryIntervalDefault
	} else {
		interval, err := time.ParseDuration(interval)
		if err != nil {
			logger.With("default", retryIntervalDefault).Info(invalidRetryInterval)
			retryInterval = retryIntervalDefault
		} else {
			retryInterval = interval
		}
	}

	maxRetry := os.Getenv("MAX_RETRY")
	if maxRetry == "" {
		logger.With("default", retryCountDefault).Info("MAX_RETRY not set")
		retries = retryCountDefault
	} else {
		max, err := strconv.Atoi(maxRetry)
		if err != nil {
			logger.With("default", retryCountDefault).Info(invalidMaxRetry)
			retries = retryCountDefault
		} else {
			retries = max
		}
	}
}
