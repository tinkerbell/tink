package main

import (
	"os"
	"strconv"
	"time"

	"github.com/tinkerbell/tink/client"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	retryIntervalDefault = 3
	retryCountDefault    = 3
)

var (
	rClient       pb.WorkflowSvcClient
	retryInterval time.Duration
	retries       int
)

var logger = logrus.New()

func main() {
	setupRetry()
	client.Setup()
	conn, err := tryClientConnection()
	if err != nil {
		logger.Fatalln(err)
	}
	rClient = pb.NewWorkflowSvcClient(conn)
	err = processWorkflowActions(rClient)
	if err != nil {
		logger.Errorln("Worker Finished with error", err)
	}
}

func tryClientConnection() (*grpc.ClientConn, error) {
	var err error
	for r := 1; r <= retries; r++ {
		c, e := client.GetConnection()
		if e != nil {
			err = e
			logger.Errorln(err)
			logger.Errorf("Retrying after %v seconds", retryInterval)
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
		logger.Infof("RETRY_INTERVAL not set. Using default, %d seconds\n", retryIntervalDefault)
		retryInterval = retryIntervalDefault
	} else {
		interval, err := time.ParseDuration(interval)
		if err != nil {
			logger.Warningf("Invalid RETRY_INTERVAL set. Using default, %d seconds.\n", retryIntervalDefault)
			retryInterval = retryIntervalDefault
		} else {
			retryInterval = interval
		}
	}

	maxRetry := os.Getenv("MAX_RETRY")
	if maxRetry == "" {
		logger.Infof("MAX_RETRY not set. Using default, %d retries.\n", retryCountDefault)
		retries = retryCountDefault
	} else {
		max, err := strconv.Atoi(maxRetry)
		if err != nil {
			logger.Warningf("Invalid MAX_RETRY set. Using default, %d retries.\n", retryCountDefault)
			retries = retryCountDefault
		} else {
			retries = max
		}
	}
}
