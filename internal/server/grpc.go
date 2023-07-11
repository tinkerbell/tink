package server

import (
	"context"

	workflowproto "github.com/tinkerbell/tink/internal/proto/workflow/v2"
)

type GRPCServer struct {
	workflowproto.UnimplementedWorkflowServiceServer
}

// GetWorkflows creates a stream that will receive workflows intended for the agent identified
// by the GetWorkflowsRequest.agent_id.
func (s *GRPCServer) GetWorkflows(req *workflowproto.GetWorkflowsRequest, stream workflowproto.WorkflowService_GetWorkflowsServer) error {
	return nil
}

// PublishEvent publishes a workflow event.
func (s *GRPCServer) PublishEvent(context.Context, *workflowproto.PublishEventRequest) (*workflowproto.PublishEventResponse, error) {
	return nil, nil
}
