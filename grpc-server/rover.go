package grpcserver

import (
	"context"

	exec "github.com/packethost/rover/executor"
	pb "github.com/packethost/rover/protos/workflow"
)

// GetWorkflowContexts implements rover.GetWorkflowContexts
func (s *server) GetWorkflowContexts(context context.Context, req *pb.WorkflowContextRequest) (*pb.WorkflowContextList, error) {
	return exec.GetWorkflowContexts(context, req, s.db)
}

// GetWorkflowActions implements rover.GetWorkflowActions
func (s *server) GetWorkflowActions(context context.Context, req *pb.WorkflowActionsRequest) (*pb.WorkflowActionList, error) {
	return exec.GetWorkflowActions(context, req, s.db)
}

// ReportActionStatus implements rover.ReportActionStatus
func (s *server) ReportActionStatus(context context.Context, req *pb.WorkflowActionStatus) (*pb.Empty, error) {
	return exec.ReportActionStatus(context, req, s.db)
}

// Update Workflow Ephemeral Data
func (s *server) UpdateWorkflowData(context context.Context, req *pb.UpdateWorkflowDataRequest) (*pb.Empty, error) {
	return exec.UpdateWorkflowData(context, req, s.db)
}

// Get Workflow Ephemeral Data
func (s *server) GetWorkflowData(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	return exec.GetWorkflowData(context, req, s.db)
}
