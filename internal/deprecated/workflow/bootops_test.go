package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestHandleExistingJob(t *testing.T) {
	tests := map[string]struct {
		workflow     *v1alpha1.Workflow
		wantWorkflow *v1alpha1.Workflow
		wantResult   reconcile.Result
		job          *rufio.Job
	}{
		"existing job deleted": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "machine1",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						ExistingJobDeleted: false,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "machine1",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						ExistingJobDeleted: false,
					},
				},
			},
			wantResult: reconcile.Result{Requeue: true},
			job: &rufio.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "tink-controller-machine1-one-time-netboot",
					Namespace: "default",
				},
			},
		},
		"no existing job": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "machine1",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "machine1",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						ExistingJobDeleted: true,
					},
				},
			},
			wantResult: reconcile.Result{Requeue: true},
		},
		"existing job already deleted": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "machine1",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						ExistingJobDeleted: true,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "machine1",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						ExistingJobDeleted: true,
					},
				},
			},
			wantResult: reconcile.Result{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			runtimescheme := runtime.NewScheme()
			rufio.AddToScheme(runtimescheme)
			v1alpha1.AddToScheme(runtimescheme)
			clientBulider := GetFakeClientBuilder().WithScheme(runtimescheme)
			if tc.job != nil {
				clientBulider.WithRuntimeObjects(tc.job)
			}
			cc := clientBulider.Build()

			r, err := handleExistingJob(context.Background(), cc, tc.workflow)
			if err != nil {
				t.Fatalf("handleExistingJob() err = %v, want nil", err)
			}
			if diff := cmp.Diff(tc.wantResult, r); diff != "" {
				t.Errorf("handleExistingJob() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantWorkflow, tc.workflow); diff != "" {
				t.Errorf("handleExistingJob() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandleJobCreation(t *testing.T) {
	uid := uuid.NewUUID()
	tests := map[string]struct {
		workflow     *v1alpha1.Workflow
		wantWorkflow *v1alpha1.Workflow
		wantResult   reconcile.Result
		job          *rufio.Job
	}{
		"creation already done": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						UID:                uid,
						ExistingJobDeleted: true,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						UID:                uid,
						ExistingJobDeleted: true,
					},
				},
			},
			wantResult: reconcile.Result{},
		},
		"create new job": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "machine1",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						ExistingJobDeleted: true,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "machine1",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						ExistingJobDeleted: true,
					},
					Conditions: []v1alpha1.WorkflowCondition{
						{Type: v1alpha1.NetbootJobSetupComplete, Status: v1.ConditionTrue, Reason: "Created", Message: "job created"},
					},
				},
			},
			wantResult: reconcile.Result{Requeue: true},
			job: &rufio.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:            "tink-controller-machine1-one-time-netboot",
					Namespace:       "default",
					ResourceVersion: "1",
					Labels: map[string]string{
						"tink-controller-auto-created": "true",
					},
					Annotations: map[string]string{
						"tink-controller-auto-created": "true",
					},
				},
				Spec: rufio.JobSpec{
					MachineRef: rufio.MachineRef{
						Name:      "machine1",
						Namespace: "default",
					},
					Tasks: []rufio.Action{
						{PowerAction: ptr.To(rufio.PowerHardOff)},
						{OneTimeBootDeviceAction: &rufio.OneTimeBootDeviceAction{Devices: []rufio.BootDevice{rufio.PXE}}},
						{PowerAction: ptr.To(rufio.PowerOn)},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			runtimescheme := runtime.NewScheme()
			rufio.AddToScheme(runtimescheme)
			v1alpha1.AddToScheme(runtimescheme)
			clientBulider := GetFakeClientBuilder().WithScheme(runtimescheme)
			clientBulider.WithRuntimeObjects(&v1alpha1.Hardware{
				ObjectMeta: v1.ObjectMeta{
					Name:      "machine1",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					BMCRef: &corev1.TypedLocalObjectReference{Name: "machine1"},
				},
			})
			cc := clientBulider.Build()

			r, err := handleJobCreation(context.Background(), cc, tc.workflow)
			if err != nil {
				t.Fatalf("handleJobCreation() err = %v, want nil", err)
			}
			if diff := cmp.Diff(tc.wantResult, r); diff != "" {
				t.Errorf("handleJobCreation() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantWorkflow, tc.workflow, cmpopts.IgnoreFields(v1alpha1.WorkflowCondition{}, "Time")); diff != "" {
				t.Errorf("handleJobCreation() mismatch (-want +got):\n%s", diff)
			}
			// check if the job is created
			if tc.job != nil {
				job := &rufio.Job{}
				if err := cc.Get(context.Background(), client.ObjectKey{Name: tc.job.Name, Namespace: tc.job.Namespace}, job); err != nil {
					t.Fatalf("handleJobCreation() job not created: %v", err)
				}
				if diff := cmp.Diff(tc.job, job); diff != "" {
					t.Errorf("handleJobCreation() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestHandleJobComplete(t *testing.T) {
	uid := uuid.NewUUID()
	tests := map[string]struct {
		workflow     *v1alpha1.Workflow
		wantWorkflow *v1alpha1.Workflow
		wantResult   reconcile.Result
		job          *rufio.Job
		shouldError  bool
	}{
		"status for existing job complete": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						Complete: true,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						Complete: true,
					},
				},
			},
			wantResult: reconcile.Result{},
		},
		"existing job not complete": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{HardwareRef: "machine1"},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						Complete:           false,
						UID:                uid,
						ExistingJobDeleted: true,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{HardwareRef: "machine1"},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						Complete:           false,
						UID:                uid,
						ExistingJobDeleted: true,
					},
					Conditions: []v1alpha1.WorkflowCondition{
						{Type: v1alpha1.NetbootJobRunning, Status: v1.ConditionTrue, Reason: "Running", Message: "one time netboot job running"},
					},
				},
			},
			wantResult: reconcile.Result{RequeueAfter: 5 * time.Second},
			job: &rufio.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "tink-controller-machine1-one-time-netboot",
					Namespace: "default",
				},
				Status: rufio.JobStatus{
					Conditions: []rufio.JobCondition{
						{Type: rufio.JobRunning, Status: rufio.ConditionTrue},
					},
				},
			},
		},
		"existing job failed": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{HardwareRef: "machine1"},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						Complete:           false,
						UID:                uid,
						ExistingJobDeleted: true,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{HardwareRef: "machine1"},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						Complete:           false,
						UID:                uid,
						ExistingJobDeleted: true,
					},
					Conditions: []v1alpha1.WorkflowCondition{
						{Type: v1alpha1.NetbootJobFailed, Status: v1.ConditionTrue, Reason: "Error", Message: "one time netboot job failed"},
					},
				},
			},
			wantResult: reconcile.Result{},
			job: &rufio.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "tink-controller-machine1-one-time-netboot",
					Namespace: "default",
				},
				Status: rufio.JobStatus{
					Conditions: []rufio.JobCondition{
						{Type: rufio.JobFailed, Status: rufio.ConditionTrue},
					},
				},
			},
			shouldError: true,
		},
		"existing job completed": {
			workflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{HardwareRef: "machine1"},
				Status: v1alpha1.WorkflowStatus{
					Job: v1alpha1.JobStatus{
						Complete:           false,
						UID:                uid,
						ExistingJobDeleted: true,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:      "workflow1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{HardwareRef: "machine1"},
				Status: v1alpha1.WorkflowStatus{
					State: v1alpha1.WorkflowStatePending,
					Job: v1alpha1.JobStatus{
						Complete:           true,
						UID:                uid,
						ExistingJobDeleted: true,
					},
					Conditions: []v1alpha1.WorkflowCondition{
						{Type: v1alpha1.NetbootJobComplete, Status: v1.ConditionTrue, Reason: "Complete", Message: "one time netboot job completed"},
					},
				},
			},
			wantResult: reconcile.Result{Requeue: true},
			job: &rufio.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "tink-controller-machine1-one-time-netboot",
					Namespace: "default",
				},
				Status: rufio.JobStatus{
					Conditions: []rufio.JobCondition{
						{Type: rufio.JobCompleted, Status: rufio.ConditionTrue},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			runtimescheme := runtime.NewScheme()
			rufio.AddToScheme(runtimescheme)
			v1alpha1.AddToScheme(runtimescheme)
			clientBulider := GetFakeClientBuilder().WithScheme(runtimescheme)
			if tc.job != nil {
				clientBulider.WithRuntimeObjects(tc.job)
			}
			cc := clientBulider.Build()

			r, err := handleJobComplete(context.Background(), cc, tc.workflow)
			if err != nil && !tc.shouldError {
				t.Fatalf("handleJobComplete() err = %v, want nil", err)
			}
			if diff := cmp.Diff(tc.wantResult, r); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantWorkflow, tc.workflow, cmpopts.IgnoreFields(v1alpha1.WorkflowCondition{}, "Time")); diff != "" {
				t.Errorf("workflow mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
