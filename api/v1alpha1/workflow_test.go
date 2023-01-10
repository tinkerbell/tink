package v1alpha1

import (
	"testing"
	"time"

	"github.com/tinkerbell/tink/internal/testtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var TestNow = testtime.NewFrozenTimeUnix(1637361793)

func TestWorkflowTinkID(t *testing.T) {
	id := "d2c26e20-97e0-449c-b665-61efa7373f47"
	cases := []struct {
		name      string
		input     *Workflow
		want      string
		overwrite string
	}{
		{
			"Already set",
			&Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
					Annotations: map[string]string{
						WorkflowIDAnnotation: id,
					},
				},
			},
			id,
			"",
		},
		{
			"nil annotations",
			&Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "debian",
					Namespace:   "default",
					Annotations: nil,
				},
			},
			"",
			"abc",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.input.TinkID() != tc.want {
				t.Errorf("Got unexpected ID: got %v, wanted %v", tc.input.TinkID(), tc.want)
			}

			tc.input.SetTinkID(tc.overwrite)

			if tc.input.TinkID() != tc.overwrite {
				t.Errorf("Got unexpected ID: got %v, wanted %v", tc.input.TinkID(), tc.overwrite)
			}
		})
	}
}

func TestGetStartTime(t *testing.T) {
	cases := []struct {
		name  string
		input *Workflow
		want  *metav1.Time
	}{
		{
			"Empty wflow",
			&Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
			},
			nil,
		},
		{
			"Running workflow",
			&Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: WorkflowSpec{},
				Status: WorkflowStatus{
					State:         WorkflowStateRunning,
					GlobalTimeout: 600,
					Tasks: []Task{
						{
							Name:       "os-installation",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    WorkflowStateSuccess,
									StartedAt: TestNow.MetaV1Now(),
									Seconds:   20,
								},
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    WorkflowStateRunning,
									StartedAt: TestNow.MetaV1AfterSec(21),
								},
							},
						},
					},
				},
			},
			TestNow.MetaV1Now(),
		},
		{
			"pending without a start time",
			&Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: WorkflowSpec{},
				Status: WorkflowStatus{
					State:         WorkflowStatePending,
					GlobalTimeout: 600,
					Tasks: []Task{
						{
							Name:       "os-installation",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    WorkflowStatePending,
									StartedAt: nil,
								},
							},
						},
					},
				},
			},
			nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.GetStartTime()
			if got == nil && tc.want == nil {
				return
			}
			if !got.Time.Equal(tc.want.Time) {
				t.Errorf("Got time %s, wanted %s", got.Format(time.RFC1123), tc.want.Time.Format(time.RFC1123))
			}
		})
	}
}

func TestWorkflowMethods(t *testing.T) {
	cases := []struct {
		name string
		wf   *Workflow
		want taskInfo
	}{
		{
			"Empty wflow",
			&Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
			},
			taskInfo{},
		},
		{
			"invalid workflow",
			&Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: WorkflowSpec{},
				Status: WorkflowStatus{
					State:         WorkflowStateRunning,
					GlobalTimeout: 600,
					Tasks: []Task{
						{
							Name: "empty task",
							// WorkerAddr: "", // intentionally not set
							Actions: []Action{
								{
									Name:   "empty action",
									Status: WorkflowStateFailed,
								},
							},
						},
						{
							Name:       "os-installation",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    WorkflowStateSuccess,
									StartedAt: TestNow.MetaV1Now(),
									Seconds:   20,
								},
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    WorkflowStateRunning,
									StartedAt: TestNow.MetaV1AfterSec(21),
								},
							},
						},
					},
				},
			},
			taskInfo{
				TotalNumberOfActions: 3,
				CurrentTaskIndex:     0,
				CurrentTask:          "empty task",
				CurrentWorker:        "",
				CurrentAction:        "empty action",
				CurrentActionState:   WorkflowStateFailed,
				CurrentActionIndex:   0,
			},
		},
		{
			"Running workflow",
			&Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: WorkflowSpec{},
				Status: WorkflowStatus{
					State:         WorkflowStateRunning,
					GlobalTimeout: 600,
					Tasks: []Task{
						{
							Name:       "bmc-manage",
							WorkerAddr: "pbnj",
							Actions: []Action{
								{
									Name:      "configure-pxe",
									Image:     "quay.io/tinkerbell-actions/pbnj:v1.0.0",
									Timeout:   20,
									Status:    WorkflowStateSuccess,
									StartedAt: TestNow.MetaV1BeforeSec(15),
									Seconds:   15,
								},
							},
						},
						{
							Name:       "os-installation",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    WorkflowStateSuccess,
									StartedAt: TestNow.MetaV1Now(),
									Seconds:   20,
								},
								{
									Name:    "write-file",
									Image:   "quay.io/tinkerbell-actions/writefile:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    WorkflowStateRunning,
									StartedAt: TestNow.MetaV1AfterSec(21),
								},
							},
						},
					},
				},
			},
			taskInfo{
				TotalNumberOfActions: 3,
				CurrentTaskIndex:     1,
				CurrentTask:          "os-installation",
				CurrentWorker:        "3c:ec:ef:4c:4f:54",
				CurrentAction:        "write-file",
				CurrentActionState:   WorkflowStateRunning,
				CurrentActionIndex:   2,
			},
		},
		{
			"Pending workflow",
			&Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: WorkflowSpec{},
				Status: WorkflowStatus{
					State:         WorkflowStatePending,
					GlobalTimeout: 600,
					Tasks: []Task{
						{
							Name:       "os-installation",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status: WorkflowStatePending,
								},
								{
									Name:    "write-file",
									Image:   "quay.io/tinkerbell-actions/writefile:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status: WorkflowStatePending,
								},
							},
						},
					},
				},
			},
			taskInfo{
				TotalNumberOfActions: 2,
				CurrentTaskIndex:     0,
				CurrentTask:          "os-installation",
				CurrentWorker:        "3c:ec:ef:4c:4f:54",
				CurrentAction:        "stream-debian-image",
				CurrentActionState:   WorkflowStatePending,
				CurrentActionIndex:   0,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.wf.getTaskActionInfo()
			if got != tc.want {
				t.Errorf("Got \n\t%#v\nwanted:\n\t%#v", got, tc.want)
			}
		})
	}
}
