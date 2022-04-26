package server

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/internal/tests"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	"github.com/tinkerbell/tink/protos/workflow"
)

var TestTime = tests.NewFrozenTimeUnix(1637361793)

func TestModifyWorkflowState(t *testing.T) {
	cases := []struct {
		name           string
		inputWf        *v1alpha1.Workflow
		inputWfContext *workflow.WorkflowContext
		want           *v1alpha1.Workflow
		wantErr        error
	}{
		{
			name:           "no workflow",
			inputWf:        nil,
			inputWfContext: &workflow.WorkflowContext{},
			want:           nil,
			wantErr:        errors.New("no workflow provided"),
		},
		{
			name:           "no context",
			inputWf:        &v1alpha1.Workflow{},
			inputWfContext: nil,
			want:           nil,
			wantErr:        errors.New("no workflow context provided"),
		},
		{
			name: "no task",
			inputWf: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_PENDING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Status:  "STATE_PENDING",
								},
							},
						},
					},
				},
			},
			inputWfContext: &workflow.WorkflowContext{
				WorkflowId:           "debian",
				CurrentWorker:        "machine-mac-1",
				CurrentTask:          "power-on",
				CurrentAction:        "power-on-bmc",
				CurrentActionIndex:   0,
				CurrentActionState:   workflow.State_STATE_RUNNING,
				TotalNumberOfActions: 1,
			},
			want:    nil,
			wantErr: errors.New("task not found"),
		},
		{
			name: "no action found",
			inputWf: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_PENDING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Status:  "STATE_PENDING",
								},
							},
						},
					},
				},
			},
			inputWfContext: &workflow.WorkflowContext{
				CurrentWorker:        "machine-mac-1",
				CurrentTask:          "provision",
				CurrentAction:        "power-on-bmc",
				CurrentActionIndex:   0,
				CurrentActionState:   workflow.State_STATE_RUNNING,
				TotalNumberOfActions: 1,
			},
			want:    nil,
			wantErr: errors.New("action not found"),
		},
		{
			name: "running task",
			inputWf: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_PENDING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Status:  "STATE_PENDING",
								},
							},
						},
					},
				},
			},
			inputWfContext: &workflow.WorkflowContext{
				CurrentWorker:        "machine-mac-1",
				CurrentTask:          "provision",
				CurrentAction:        "stream",
				CurrentActionIndex:   0,
				CurrentActionState:   workflow.State_STATE_RUNNING,
				TotalNumberOfActions: 1,
			},
			want: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:      "stream",
									Image:     "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout:   300,
									Status:    "STATE_RUNNING",
									StartedAt: TestTime.MetaV1Now(),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "timed out task",
			inputWf: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:      "stream",
									Image:     "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout:   300,
									Status:    "STATE_RUNNING",
									StartedAt: TestTime.MetaV1Before(time.Second * 301),
								},
							},
						},
					},
				},
			},
			inputWfContext: &workflow.WorkflowContext{
				CurrentWorker:        "machine-mac-1",
				CurrentTask:          "provision",
				CurrentAction:        "stream",
				CurrentActionIndex:   0,
				CurrentActionState:   workflow.State_STATE_TIMEOUT,
				TotalNumberOfActions: 1,
			},
			want: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_TIMEOUT",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:      "stream",
									Image:     "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout:   300,
									Status:    "STATE_TIMEOUT",
									StartedAt: TestTime.MetaV1Before(time.Second * 301),
									Seconds:   301,
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "successful task",
			inputWf: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:      "stream",
									Image:     "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout:   300,
									Status:    "STATE_RUNNING",
									StartedAt: TestTime.MetaV1Before(time.Second * 30),
								},
								{
									Name:    "kexec",
									Image:   "quay.io/tinkerbell-actions/kexec:v1.0.0",
									Timeout: 5,
									Status:  "STATE_PENDING",
								},
							},
						},
					},
				},
			},
			inputWfContext: &workflow.WorkflowContext{
				CurrentWorker:        "machine-mac-1",
				CurrentTask:          "provision",
				CurrentAction:        "stream",
				CurrentActionIndex:   0,
				CurrentActionState:   workflow.State_STATE_SUCCESS,
				TotalNumberOfActions: 2,
			},
			want: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:      "stream",
									Image:     "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout:   300,
									Status:    "STATE_SUCCESS",
									StartedAt: TestTime.MetaV1Before(time.Second * 30),
									Seconds:   30,
								},
								{
									Name:    "kexec",
									Image:   "quay.io/tinkerbell-actions/kexec:v1.0.0",
									Timeout: 5,
									Status:  "STATE_PENDING",
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "successful last task",
			inputWf: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:      "stream",
									Image:     "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout:   300,
									Status:    "STATE_SUCCESS",
									StartedAt: TestTime.MetaV1Before(time.Second * 30),
									Seconds:   27,
								},
								{
									Name:    "kexec",
									Image:   "quay.io/tinkerbell-actions/kexec:v1.0.0",
									Timeout: 5,
									Status:  "STATE_RUNNING",
								},
							},
						},
					},
				},
			},
			inputWfContext: &workflow.WorkflowContext{
				CurrentWorker:        "machine-mac-1",
				CurrentTask:          "provision",
				CurrentAction:        "kexec",
				CurrentActionIndex:   1,
				CurrentActionState:   workflow.State_STATE_SUCCESS,
				TotalNumberOfActions: 2,
			},
			want: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_SUCCESS",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "machine-mac-1",
							Actions: []v1alpha1.Action{
								{
									Name:      "stream",
									Image:     "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout:   300,
									Status:    "STATE_SUCCESS",
									StartedAt: TestTime.MetaV1Before(time.Second * 30),
									Seconds:   27,
								},
								{
									Name:    "kexec",
									Image:   "quay.io/tinkerbell-actions/kexec:v1.0.0",
									Timeout: 5,
									Status:  "STATE_SUCCESS",
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := &KubernetesBackedServer{
				logger:     log.Test(t, "TestModifyWorkflowState"),
				ClientFunc: nil,
				namespace:  "default",
				nowFunc:    TestTime.Now,
			}
			gotErr := server.modifyWorkflowState(tc.inputWf, tc.inputWfContext)
			tests.CompareErrors(t, gotErr, tc.wantErr)
			if tc.want == nil {
				return
			}

			if diff := cmp.Diff(tc.inputWf, tc.want); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}
