package server

import "github.com/tinkerbell/tink/internal/workflow"

/*
The server must be notified when an agent connects.
The server must be notified when an agent disconnects.
The server must dispatch workflows to the agent.
The server must be notified of events created by the agent.
The server must ensure a single workflow is dispatched to an agent at a time.
The server must time out workflows.
*/

type Transport interface {
	Send(workflow.Workflow) error
}

type Server struct {
	agents map[string]Transport
}

func (s *Server) RegisterAgent(agentID string, trnsport Transport) error {
	return nil
}

func (s *Server) UnregisterAgent(agentID string) error {
	return nil
}
