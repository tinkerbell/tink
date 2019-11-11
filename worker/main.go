package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/packethost/rover/client"
	pb "github.com/packethost/rover/protos/rover"
	"google.golang.org/grpc"
)

const (
	retryIntervalDefault = 3
	retryCountDefault    = 3
)

var (
	rClient       pb.RoverClient
	retryInterval time.Duration
	retries       int
)

func main() {
	setupRetry()
	conn, err := tryClientConnection()
	if err != nil {
		log.Fatalln(err)
	}
	rClient = pb.NewRoverClient(conn)
	err = initializeWorker(rClient)
	if err != nil {
		log.Fatalln(err)
	}
}

func tryClientConnection() (*grpc.ClientConn, error) {
	var err error
	for r := 1; r <= retries; r++ {
		c, e := client.GetConnection()
		if e != nil {
			err = e
			log.Println(err)
			log.Printf("Retrying after %v seconds", retryInterval)
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
		log.Printf("RETRY_INTERVAL not set. Using default, %d seconds\n", retryIntervalDefault)
		retryInterval = retryIntervalDefault
	} else {
		interval, err := time.ParseDuration(interval)
		if err != nil {
			log.Printf("Invalid RETRY_INTERVAL set. Using default, %d seconds.\n", retryIntervalDefault)
			retryInterval = retryIntervalDefault
		} else {
			retryInterval = interval
		}
	}

	maxRetry := os.Getenv("MAX_RETRY")
	if maxRetry == "" {
		log.Printf("MAX_RETRY not set. Using default, %d retries.\n", retryCountDefault)
		retries = retryCountDefault
	} else {
		max, err := strconv.Atoi(maxRetry)
		if err != nil {
			log.Printf("Invalid MAX_RETRY set. Using default, %d retries.\n", retryCountDefault)
			retries = retryCountDefault
		} else {
			retries = max
		}
	}
}
