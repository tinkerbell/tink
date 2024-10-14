package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/deprecated/workflow/journal"
	"github.com/tinkerbell/tink/internal/ptr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestPrepareWorkflow(t *testing.T) {
	tests := map[string]struct {
		wantResult   reconcile.Result
		wantError    bool
		hardware     *v1alpha1.Hardware
		wantHardware *v1alpha1.Hardware
		workflow     *v1alpha1.Workflow
		wantWorkflow *v1alpha1.Workflow
		job          *rufio.Job
	}{
		"nothing to do": {
			wantResult:   reconcile.Result{},
			hardware:     &v1alpha1.Hardware{},
			wantHardware: &v1alpha1.Hardware{},
			workflow:     &v1alpha1.Workflow{},
			wantWorkflow: &v1alpha1.Workflow{},
		},
		"toggle allowPXE": {
			wantResult: reconcile.Result{},
			hardware: &v1alpha1.Hardware{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-hardware",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							Netboot: &v1alpha1.Netboot{
								AllowPXE: ptr.Bool(false),
							},
						},
					},
				},
			},
			wantHardware: &v1alpha1.Hardware{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-hardware",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							Netboot: &v1alpha1.Netboot{
								AllowPXE: ptr.Bool(true),
							},
						},
					},
				},
			},
			workflow: &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workflow",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "test-hardware",
					BootOptions: v1alpha1.BootOptions{
						ToggleAllowNetboot: true,
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					BootOptions: v1alpha1.BootOptionsStatus{
						AllowNetboot: v1alpha1.AllowNetbootStatus{
							ToggledTrue: true,
						},
					},
					Conditions: []v1alpha1.WorkflowCondition{
						{
							Type:    v1alpha1.ToggleAllowNetbootTrue,
							Status:  metav1.ConditionTrue,
							Reason:  "Complete",
							Message: "set allowPXE to true",
						},
					},
				},
			},
		},
		"boot mode netboot": {
			wantResult: reconcile.Result{Requeue: true},
			hardware: &v1alpha1.Hardware{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-hardware",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					BMCRef: &v1.TypedLocalObjectReference{
						Name: "test-bmc",
						Kind: "machine.bmc.tinkerbell.org",
					},
				},
			},
			wantHardware: &v1alpha1.Hardware{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-hardware",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					BMCRef: &v1.TypedLocalObjectReference{
						Name: "test-bmc",
						Kind: "machine.bmc.tinkerbell.org",
					},
				},
			},
			workflow: &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workflow",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "test-hardware",
					BootOptions: v1alpha1.BootOptions{
						BootMode: "netboot",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					BootOptions: v1alpha1.BootOptionsStatus{
						Jobs: map[string]v1alpha1.JobStatus{},
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					BootOptions: v1alpha1.BootOptionsStatus{
						Jobs: map[string]v1alpha1.JobStatus{
							fmt.Sprintf("%s-test-workflow", jobNameNetboot): {ExistingJobDeleted: true},
						},
					},
				},
			},
		},
		"boot mode iso": {
			wantResult: reconcile.Result{Requeue: true},
			hardware: &v1alpha1.Hardware{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-hardware",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					BMCRef: &v1.TypedLocalObjectReference{
						Name: "test-bmc",
						Kind: "machine.bmc.tinkerbell.org",
					},
				},
			},
			wantHardware: &v1alpha1.Hardware{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-hardware",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					BMCRef: &v1.TypedLocalObjectReference{
						Name: "test-bmc",
						Kind: "machine.bmc.tinkerbell.org",
					},
				},
			},
			workflow: &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workflow",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					HardwareRef: "test-hardware",
					BootOptions: v1alpha1.BootOptions{
						BootMode: "iso",
						ISOURL:   "http://example.com",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					BootOptions: v1alpha1.BootOptionsStatus{
						Jobs: map[string]v1alpha1.JobStatus{},
					},
				},
			},
			wantWorkflow: &v1alpha1.Workflow{
				Status: v1alpha1.WorkflowStatus{
					BootOptions: v1alpha1.BootOptionsStatus{
						Jobs: map[string]v1alpha1.JobStatus{
							fmt.Sprintf("%s-test-workflow", jobNameISOMount): {ExistingJobDeleted: true},
						},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			rufio.AddToScheme(scheme)
			v1alpha1.AddToScheme(scheme)
			ro := []runtime.Object{}
			if tc.hardware != nil {
				ro = append(ro, tc.hardware)
			}
			if tc.workflow != nil {
				ro = append(ro, tc.workflow)
			}
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ro...)
			if tc.job != nil {
				clientBuilder.WithRuntimeObjects(tc.job)
			}
			s := &state{
				workflow: tc.workflow,
				client:   clientBuilder.Build(),
			}
			ctx := context.Background()
			ctx = journal.New(ctx)
			result, err := s.prepareWorkflow(ctx)
			if (err != nil) != tc.wantError {
				t.Errorf("expected error: %v, got: %v", tc.wantError, err)
			}
			if diff := cmp.Diff(result, tc.wantResult); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
				t.Logf("journal: %v", journal.Journal(ctx))
			}

			// get the Hardware object in cluster
			gotHardware := &v1alpha1.Hardware{}
			if err := s.client.Get(ctx, types.NamespacedName{Name: tc.hardware.Name, Namespace: tc.hardware.Namespace}, gotHardware); err != nil {
				t.Fatalf("error getting hardware: %v", err)
			}
			if diff := cmp.Diff(gotHardware.Spec, tc.wantHardware.Spec); diff != "" {
				t.Errorf("unexpected hardware (-want +got):\n%s", diff)
				for _, entry := range journal.Journal(ctx) {
					t.Logf("journal: %+v", entry)
				}
			}

			if diff := cmp.Diff(tc.workflow.Status, tc.wantWorkflow.Status, cmpopts.IgnoreFields(v1alpha1.WorkflowCondition{}, "Time")); diff != "" {
				t.Errorf("unexpected workflow status (-want +got):\n%s", diff)
				for _, entry := range journal.Journal(ctx) {
					t.Logf("journal: %+v", entry)
				}
			}
		})
	}
}
