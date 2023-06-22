package cli

import (
	"fmt"

	"github.com/go-logr/zapr"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/internal/agent"
	"github.com/tinkerbell/tink/internal/agent/runtime"
	"github.com/tinkerbell/tink/internal/agent/transport"
	"github.com/tinkerbell/tink/internal/proto/workflow/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NewAgent builds a command that launches the agent component.
func NewAgent() *cobra.Command {
	var opts struct {
		AgentID        string
		TinkServerAddr string
	}

	// TODO(chrisdoherty4) Handle signals
	cmd := cobra.Command{
		Use: "tink-agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			zl, err := zap.NewProduction()
			if err != nil {
				return fmt.Errorf("init logger: %w", err)
			}
			logger := zapr.NewLogger(zl)

			rntime, err := runtime.NewDocker()
			if err != nil {
				return fmt.Errorf("create runtime: %w", err)
			}

			conn, err := grpc.DialContext(cmd.Context(), opts.TinkServerAddr)
			if err != nil {
				return fmt.Errorf("dial tink server: %w", err)
			}
			defer conn.Close()
			trnport := transport.NewGRPC(logger, workflow.NewWorkflowServiceClient(conn))

			return (&agent.Agent{
				Log:       logger,
				ID:        opts.AgentID,
				Transport: trnport,
				Runtime:   rntime,
			}).Start(cmd.Context())
		},
	}

	flgs := cmd.Flags()
	flgs.StringVar(&opts.AgentID, "agent-id", "", "An ID that uniquely identifies the agent instance")
	flgs.StringVar(&opts.TinkServerAddr, "tink-server-addr", "127.0.0.1:42113", "Tink server address")

	return &cmd
}
