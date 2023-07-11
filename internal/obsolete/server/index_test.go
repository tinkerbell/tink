package server

import (
	"reflect"
	"testing"

	"github.com/tinkerbell/tink/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestWorkflowIndexFuncs(t *testing.T) {
	cases := []struct {
		name           string
		input          client.Object
		wantStateAddrs []string
	}{
		{
			"non workflow",
			&v1alpha1.Hardware{},
			nil,
		},
		{
			"empty workflow",
			&v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State: "",
					Tasks: []v1alpha1.Task{},
				},
			},
			[]string{},
		},
		{
			"pending workflow",
			&v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State: v1alpha1.WorkflowStatePending,
					Tasks: []v1alpha1.Task{
						{
							WorkerAddr: "worker1",
						},
					},
				},
			},
			[]string{"worker1"},
		},
		{
			"running workflow",
			&v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State: v1alpha1.WorkflowStateRunning,
					Tasks: []v1alpha1.Task{
						{
							WorkerAddr: "worker1",
						},
						{
							WorkerAddr: "worker2",
						},
					},
				},
			},
			[]string{"worker1", "worker2"},
		},
		{
			"complete workflow",
			&v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State: v1alpha1.WorkflowStateSuccess,
					Tasks: []v1alpha1.Task{
						{
							WorkerAddr: "worker1",
						},
					},
				},
			},
			[]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotStateAddrs := workflowByNonTerminalStateFunc(tc.input)
			if !reflect.DeepEqual(tc.wantStateAddrs, gotStateAddrs) {
				t.Errorf("Unexpected non-terminating workflow response: wanted %#v, got %#v", tc.wantStateAddrs, gotStateAddrs)
			}
		})
	}
}
