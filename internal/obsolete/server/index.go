package server

import (
	"github.com/tinkerbell/tink/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// workflowByNonTerminalState is the index name for retrieving workflows in a non-terminal state.
const workflowByNonTerminalState = ".status.state.nonTerminalWorker"

// workflowByNonTerminalStateFunc inspects obj - which must be a Workflow - for a Pending or
// Running state. If in either Pending or Running it returns a list of worker addresses.
func workflowByNonTerminalStateFunc(obj client.Object) []string {
	wf, ok := obj.(*v1alpha1.Workflow)
	if !ok {
		return nil
	}

	resp := []string{}
	if !(wf.Status.State == v1alpha1.WorkflowStateRunning || wf.Status.State == v1alpha1.WorkflowStatePending) {
		return resp
	}
	for _, task := range wf.Status.Tasks {
		if task.WorkerAddr != "" {
			resp = append(resp, task.WorkerAddr)
		}
	}

	return resp
}
