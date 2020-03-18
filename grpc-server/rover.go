package grpcserver

import (
	"context"

	exec "github.com/packethost/tinkerbell/executor"
	pb "github.com/packethost/tinkerbell/protos/workflow"
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

// GetWorkflowMetadata returns metadata wrt to the ephemeral data of a workflow
func (s *server) GetWorkflowMetadata(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	return exec.GetWorkflowMetadata(context, req, s.db)
}

// GetWorkflowDataVersion returns the latest version of data for a workflow
func (s *server) GetWorkflowDataVersion(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	return exec.GetWorkflowDataVersion(context, req.WorkflowID, s.db)
}
