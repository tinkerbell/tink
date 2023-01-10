package workflow

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/proto"
	"github.com/tinkerbell/tink/internal/testtime"
	"google.golang.org/protobuf/testing/protocmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var convertTestTime = testtime.NewFrozenTimeUnix(1637361794)

func TestToWorkflowContext(t *testing.T) {
	cases := []struct {
		name  string
		input *v1alpha1.Workflow
		want  *proto.WorkflowContext
	}{
		{
			"nil workflow",
			nil,
			nil,
		},
		{
			"empty workflow",
			&v1alpha1.Workflow{},
			&proto.WorkflowContext{
				WorkflowId:           "",
				CurrentWorker:        "",
				CurrentTask:          "",
				CurrentAction:        "",
				CurrentActionIndex:   0,
				CurrentActionState:   0,
				TotalNumberOfActions: 0,
			},
		},
		{
			"running workflow",
			&v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wf1",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "task1",
							WorkerAddr: "worker1",
							Actions: []v1alpha1.Action{
								{
									Name:   "action1",
									Status: "STATE_SUCCESS",
								},
								{
									Name:   "action2",
									Status: "STATE_RUNNING",
								},
							},
						},
					},
				},
			},
			&proto.WorkflowContext{
				WorkflowId:           "wf1",
				CurrentWorker:        "worker1",
				CurrentTask:          "task1",
				CurrentAction:        "action2",
				CurrentActionIndex:   1,
				CurrentActionState:   proto.State_STATE_RUNNING,
				TotalNumberOfActions: 2,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ToWorkflowContext(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Unexpedted response: wanted\n\t%#v\ngot\n\t%#v", tc.want, got)
			}
		})
	}
}

func TestActionListCRDToProto(t *testing.T) {
	cases := []struct {
		name  string
		input *v1alpha1.Workflow
		want  *proto.WorkflowActionList
	}{
		{
			"nil arg",
			nil,
			nil,
		},
		{
			"empty workflow",
			&v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "wf1",
					CreationTimestamp: *convertTestTime.MetaV1Now(),
				},
				Spec:   v1alpha1.WorkflowSpec{},
				Status: v1alpha1.WorkflowStatus{},
			},
			&proto.WorkflowActionList{
				ActionList: []*proto.WorkflowAction{},
			},
		},
		{
			"full workflow",
			&v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "wf1",
					Annotations: map[string]string{
						"workflow.tinkerbell.org/id": "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
					},
					CreationTimestamp: *convertTestTime.MetaV1Now(),
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "MyCoolWorkflow",
				},
				Status: v1alpha1.WorkflowStatus{
					State: "STATE_SUCCESS",
					Tasks: []v1alpha1.Task{
						{
							Name:       "worker1",
							WorkerAddr: "00:00:53:00:53:F4",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream-debian-image",
									Timeout: 600,
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Environment: map[string]string{
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
										"COMPRESSED": "true",
										"GODEBUG":    "",
									},
									Volumes: []string{
										"/tmp/debug:/tmp/debug",
									},
								},
								{
									Name:    "kexec",
									Image:   "quay.io/tinkerbell-actions/kexec:v1.0.1",
									Timeout: 90,
									Pid:     "host",
									Environment: map[string]string{
										"FS_TYPE":      "ext4",
										"BLOCK_DEVICE": "/dev/nvme0n1p1",
									},
								},
							},
							Volumes: []string{
								"/dev:/dev",
								"/dev/console:/dev/console",
								"/lib/firmware:/lib/firmware:ro",
							},
							Environment: map[string]string{
								"GODEBUG": "http2debug=1",
								"GOGC":    "100",
							},
						},
					},
				},
			},
			&proto.WorkflowActionList{
				ActionList: []*proto.WorkflowAction{
					{
						TaskName: "worker1",
						Name:     "stream-debian-image",
						Image:    "quay.io/tinkerbell-actions/image2disk:v1.0.0",
						Timeout:  600,
						WorkerId: "00:00:53:00:53:F4",
						Environment: []string{
							"COMPRESSED=true",
							"DEST_DISK=/dev/nvme0n1",
							"GODEBUG=",
							"GOGC=100",
							"IMG_URL=http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
						},
						Volumes: []string{
							"/dev:/dev",
							"/dev/console:/dev/console",
							"/lib/firmware:/lib/firmware:ro",
							"/tmp/debug:/tmp/debug",
						},
					},
					{
						TaskName: "worker1",
						Name:     "kexec",
						Image:    "quay.io/tinkerbell-actions/kexec:v1.0.1",
						Timeout:  90,
						WorkerId: "00:00:53:00:53:F4",
						Environment: []string{
							"BLOCK_DEVICE=/dev/nvme0n1p1",
							"FS_TYPE=ext4",
							"GODEBUG=http2debug=1",
							"GOGC=100",
						},
						Pid: "host",
						Volumes: []string{
							"/dev:/dev",
							"/dev/console:/dev/console",
							"/lib/firmware:/lib/firmware:ro",
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ActionListCRDToProto(tc.input)
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}

func TestYAMLToStatus(t *testing.T) {
	cases := []struct {
		name    string
		inputWf *Workflow
		want    *v1alpha1.WorkflowStatus
	}{
		{
			"Nil workflow",
			nil,
			nil,
		},
		{
			"Full crd",
			&Workflow{
				Version:       "1",
				Name:          "debian-provision",
				ID:            "0a90fac9-b509-4aa5-b294-5944128ece81",
				GlobalTimeout: 600,
				Tasks: []Task{
					{
						Name:       "do-or-do-not-there-is-no-try",
						WorkerAddr: "00:00:53:00:53:F4",
						Actions: []Action{
							{
								Name:    "stream-image-to-disk",
								Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
								Timeout: 300,
								Volumes: []string{
									"/dev:/dev",
									"/dev/console:/dev/console",
									"/lib/firmware:/lib/firmware:ro",
									"/tmp/debug:/tmp/debug",
								},
								Environment: map[string]string{
									"COMPRESSED": "true",
									"DEST_DISK":  "/dev/nvme0n1",
									"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
								},
								Pid: "host",
							},
						},
					},
				},
			},
			&v1alpha1.WorkflowStatus{
				GlobalTimeout: 600,
				Tasks: []v1alpha1.Task{
					{
						Name:       "do-or-do-not-there-is-no-try",
						WorkerAddr: "00:00:53:00:53:F4",
						Actions: []v1alpha1.Action{
							{
								Name:    "stream-image-to-disk",
								Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
								Timeout: 300,
								Volumes: []string{
									"/dev:/dev",
									"/dev/console:/dev/console",
									"/lib/firmware:/lib/firmware:ro",
									"/tmp/debug:/tmp/debug",
								},
								Pid: "host",
								Environment: map[string]string{
									"COMPRESSED": "true",
									"DEST_DISK":  "/dev/nvme0n1",
									"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
								},
								Status: "STATE_PENDING",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := YAMLToStatus(tc.inputWf)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}
