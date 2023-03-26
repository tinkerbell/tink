package agent

import (
	"context"

	"github.com/tinkerbell/tink/internal/agent/transport"
)

// Transport is a transport mechanism for communicating workflows to the agent.
type Transport interface {
	// Start is a blocking call that starts the transport and begins retreiving workflows for the
	// given agentID. The transport should pass workflows to the WorkflowHandler. If the transport
	// needs to cancel a workflow it should cancel the context passed to the WorkflowHandler.
	Start(_ context.Context, agentID string, _ transport.WorkflowHandler) error
}
