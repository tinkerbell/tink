package server

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

var errNotImplemented = errors.New("not implemented")

// CreateWorkflow will return a not implemented error.
func (s *KubernetesBackedServer) CreateWorkflow(context.Context, *pb.CreateRequest) (*pb.CreateResponse, error) {
	return nil, errNotImplemented
}

// GetWorkflow will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflow(context.Context, *pb.GetRequest) (*pb.Workflow, error) {
	return nil, errNotImplemented
}

// DeleteWorkflow will return a not implemented error.
func (s *KubernetesBackedServer) DeleteWorkflow(context.Context, *pb.GetRequest) (*pb.Empty, error) {
	return nil, errNotImplemented
}

// ListWOrkflows will return a not implemented error.
func (s *KubernetesBackedServer) ListWorkflows(*pb.Empty, pb.WorkflowService_ListWorkflowsServer) error {
	return errNotImplemented
}

// ShowWorkflowEvents will return a not implemented error.
func (s *KubernetesBackedServer) ShowWorkflowEvents(*pb.GetRequest, pb.WorkflowService_ShowWorkflowEventsServer) error {
	return errNotImplemented
}

// GetWorkflowContext will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflowContext(context.Context, *pb.GetRequest) (*pb.WorkflowContext, error) {
	return nil, errNotImplemented
}

// GetWorkflowContextList will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflowContextList(context.Context, *pb.WorkflowContextRequest) (*pb.WorkflowContextList, error) {
	return nil, errNotImplemented
}

// GetWorkflowMetadata will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflowMetadata(context.Context, *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	return nil, errNotImplemented
}

// GetWorkflowDataVersion will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflowDataVersion(context.Context, *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	return nil, errNotImplemented
}
