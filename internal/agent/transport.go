package agent

import (
	"context"

	"github.com/tinkerbell/tink/internal/agent/transport"
)

// Transport is a transport mechanism for communicating workflows to the agent.
type Transport interface {
	// Start is a blocking call that starts the transport and begins retrieving workflows for the
	// given agentID. The transport should pass workflows to the Handler. The transport
	// should block until its told to cancel via the context.
	Start(_ context.Context, agentID string, _ transport.WorkflowHandler) error
}
