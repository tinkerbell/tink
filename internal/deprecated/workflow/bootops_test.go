package workflow

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
			//cc := GetFakeClientBuilder().Build()
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
