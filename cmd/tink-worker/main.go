package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/tinkerbell/tink/cmd/tink-worker/client/tink"
	"github.com/tinkerbell/tink/cmd/tink-worker/cmd"
	"github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
)

func main() {
	// parse and validate command-line flags and required env vars
	flagEnvSettings, err := cmd.CollectFlagEnvSettings(os.Args[1:])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Here we retry failed connections to tink-server using a randomized interval between attempts.
	// tink-worker is a daemon process, so we do not exit here.
	var conn *grpc.ClientConn
	for {
		creds, err := tink.ObtainServerCreds(flagEnvSettings.TinkServerURL)
		if err == nil {
			conn, err = tink.EstablishServerConnection(flagEnvSettings.TinkServerGRPCAuthority, creds)
			if err == nil {
				break
			}
			fmt.Printf("Error establishing gPRC connection to tink-server at %s: %v:", flagEnvSettings.TinkServerGRPCAuthority, err)
		} else {
			fmt.Printf("Error obtaining server creds from %s: %v", flagEnvSettings.TinkServerURL, err)
		}

		// sleep a randomized amount of time before reconnecting to avoid thundering herds
		// TODO: we may want to make this configurable via a cmdline option?
		rand.Seed(time.Now().UnixNano())
		s := rand.Intn(120) + 1 // 2 minutes (120 seconds), plus one second to avoid a possible zero sleep time
		fmt.Printf("Sleeping %d seconds before attempting to re-connect...\n", s)
		time.Sleep(time.Duration(s) * time.Second)
	}

	workflowClient := workflow.NewWorkflowServiceClient(conn)

	// TODO: this is just here during developent to prevent linter issues, remove this
	if workflowClient == nil {
		fmt.Println("We already check conn so there's no way these could actually be nil")
		os.Exit(1)
	}

	// Remove me: this is just here during developent to prevent linter issues
	fmt.Printf("Version flag is %v\n", flagEnvSettings.Version)
}
