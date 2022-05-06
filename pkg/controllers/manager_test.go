package controllers

import (
	"reflect"
	"testing"

	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestWorkflowIndexFuncs(t *testing.T) {
	cases := []struct {
		name           string
		input          client.Object
		wantAddrs      []string
		wantStateAddrs []string
		wantStates     []string
	}{
		{
			"non workflow",
			&v1alpha1.Hardware{},
			nil,
			nil,
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
			[]string{},
			[]string{""},
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
			[]string{"worker1"},
			[]string{string(v1alpha1.WorkflowStatePending)},
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
			[]string{"worker1", "worker2"},
			[]string{string(v1alpha1.WorkflowStateRunning)},
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
			[]string{"worker1"},
			[]string{},
			[]string{string(v1alpha1.WorkflowStateSuccess)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotAddr := WorkflowWorkerAddrIndexFunc(tc.input)
			if !reflect.DeepEqual(tc.wantAddrs, gotAddr) {
				t.Errorf("Unexpected wokerAddr response: wanted %#v, got %#v", tc.wantAddrs, gotAddr)
			}
			gotStateAddrs := WorkflowWorkerNonTerminalStateIndexFunc(tc.input)
			if !reflect.DeepEqual(tc.wantStateAddrs, gotStateAddrs) {
				t.Errorf("Unexpected non-terminating workflow response: wanted %#v, got %#v", tc.wantStateAddrs, gotStateAddrs)
			}
			gotStates := WorkflowStateIndexFunc(tc.input)
			if !reflect.DeepEqual(tc.wantStates, gotStates) {
				t.Errorf("Unexpected workflow state response: wanted %#v, got %#v", tc.wantStates, gotStates)
			}
		})
	}
}

func TestHardwareIndexFunc(t *testing.T) {
	cases := []struct {
		name    string
		input   client.Object
		wantMac []string
		wantIP  []string
	}{
		{
			"non hardware",
			&v1alpha1.Workflow{},
			nil,
			nil,
		},
		{
			"empty hardware dhcp",
			&v1alpha1.Hardware{
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							DHCP: nil,
						},
					},
				},
			},
			[]string{},
			[]string{},
		},
		{
			"blank mac and ip",
			&v1alpha1.Hardware{
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							DHCP: &v1alpha1.DHCP{
								MAC: "",
								IP: &v1alpha1.IP{
									Address: "",
								},
							},
						},
					},
				},
			},
			[]string{},
			[]string{},
		},
		{
			"single ip and mac",
			&v1alpha1.Hardware{
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							DHCP: &v1alpha1.DHCP{
								MAC: "3c:ec:ef:4c:4f:54",
								IP: &v1alpha1.IP{
									Address: "172.16.10.100",
								},
							},
						},
					},
				},
			},
			[]string{"3c:ec:ef:4c:4f:54"},
			[]string{"172.16.10.100"},
		},
		{
			"double ip and mac",
			&v1alpha1.Hardware{
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							DHCP: &v1alpha1.DHCP{
								MAC: "3c:ec:ef:4c:4f:54",
								IP: &v1alpha1.IP{
									Address: "172.16.10.100",
								},
							},
						},
						{
							DHCP: &v1alpha1.DHCP{
								MAC: "3c:ec:ef:4c:4f:55",
								IP: &v1alpha1.IP{
									Address: "172.16.10.101",
								},
							},
						},
					},
				},
			},
			[]string{"3c:ec:ef:4c:4f:54", "3c:ec:ef:4c:4f:55"},
			[]string{"172.16.10.100", "172.16.10.101"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMac := HardwareMacIndexFunc(tc.input)
			if !reflect.DeepEqual(tc.wantMac, gotMac) {
				t.Errorf("Unexpected response: wanted %#v, got %#v", tc.wantMac, gotMac)
			}
			gotIPs := HardwareIPIndexFunc(tc.input)
			if !reflect.DeepEqual(tc.wantIP, gotIPs) {
				t.Errorf("Unexpected response: wanted %#v, got %#v", tc.wantIP, gotIPs)
			}
		})
	}
}
