package v1alpha1

import (
	"testing"
	"time"

	"github.com/tinkerbell/tink/pkg/internal/tests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var TestNow = tests.NewFrozenTimeUnix(1637361793)

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
					State:         "STATE_RUNNING",
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
									Status:    "STATE_SUCCESS",
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
									Status:    "STATE_RUNNING",
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
					State:         "STATE_PENDING",
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
									Status:    "STATE_PENDING",
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
