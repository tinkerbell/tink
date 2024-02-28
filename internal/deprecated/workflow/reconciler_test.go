package workflow

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/testtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var runtimescheme = runtime.NewScheme()

// TestTime is a static time that can be used for testing.
var TestTime = testtime.NewFrozenTimeUnix(1637361793)

func init() {
	_ = clientgoscheme.AddToScheme(runtimescheme)
	_ = v1alpha1.AddToScheme(runtimescheme)
}

func GetFakeClientBuilder() *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(
		runtimescheme,
	).WithRuntimeObjects(
		&v1alpha1.Hardware{}, &v1alpha1.Template{}, &v1alpha1.Workflow{},
	)
}

var minimalTemplate = `version: "0.1"
name: debian
global_timeout: 1800
tasks:
  - name: "os-installation"
    worker: "{{.device_1}}"
    volumes:
      - /dev:/dev
      - /dev/console:/dev/console
      - /lib/firmware:/lib/firmware:ro
    actions:
      - name: "stream-debian-image"
        image: quay.io/tinkerbell-actions/image2disk:v1.0.0
        timeout: 600
        environment:
          DEST_DISK: /dev/nvme0n1
          # Hegel IP
          IMG_URL: "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz"
          COMPRESSED: true`

var templateWithDiskTemplate = `version: "0.1"
name: debian
global_timeout: 1800
tasks:
  - name: "os-installation"
    worker: "{{.device_1}}"
    volumes:
      - /dev:/dev
      - /dev/console:/dev/console
      - /lib/firmware:/lib/firmware:ro
    actions:
      - name: "stream-debian-image"
        image: quay.io/tinkerbell-actions/image2disk:v1.0.0
        timeout: 600
        environment:
          DEST_DISK: {{ index .Hardware.Disks 0 }}
          # Hegel IP
          IMG_URL: "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz"
          COMPRESSED: true`

func TestReconcile(t *testing.T) {
	cases := []struct {
		name         string
		seedTemplate *v1alpha1.Template
		seedWorkflow *v1alpha1.Workflow
		seedHardware *v1alpha1.Hardware
		req          reconcile.Request
		want         reconcile.Result
		wantWflow    *v1alpha1.Workflow
		wantErr      error
	}{
		{
			name: "DoesNotExist",
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "notreal",
					Namespace: "default",
				},
			},
			want: reconcile.Result{},
			wantWflow: &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "999",
				},
			},
			wantErr: nil,
		},
		{
			name: "NewWorkflow",
			seedTemplate: &v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.TemplateSpec{
					Data: &minimalTemplate,
				},
				Status: v1alpha1.TemplateStatus{},
			},
			seedWorkflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{},
			},
			seedHardware: &v1alpha1.Hardware{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Hardware",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine1",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							Netboot: &v1alpha1.Netboot{
								AllowPXE:      &[]bool{true}[0],
								AllowWorkflow: &[]bool{true}[0],
							},
							DHCP: &v1alpha1.DHCP{
								Arch:     "x86_64",
								Hostname: "sm01",
								IP: &v1alpha1.IP{
									Address: "172.16.10.100",
									Gateway: "172.16.10.1",
									Netmask: "255.255.255.0",
								},
								LeaseTime:   86400,
								MAC:         "3c:ec:ef:4c:4f:54",
								NameServers: []string{},
								UEFI:        true,
							},
						},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "debian",
					Namespace: "default",
				},
			},
			want: reconcile.Result{},
			wantWflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1000",
					Name:            "debian",
					Namespace:       "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         v1alpha1.WorkflowStatePending,
					GlobalTimeout: 1800,
					Tasks: []v1alpha1.Task{
						{
							Name: "os-installation",

							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Volumes: []string{
								"/dev:/dev",
								"/dev/console:/dev/console",
								"/lib/firmware:/lib/firmware:ro",
							},
							Actions: []v1alpha1.Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 600,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status: v1alpha1.WorkflowStatePending,
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "MalformedWorkflow",
			seedTemplate: &v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.TemplateSpec{
					Data: &[]string{`version: "0.1"
					name: debian
global_timeout: 1800
tasks:
	- name: "os-installation"
		worker: "{{.device_1}}"`}[0],
				},
				Status: v1alpha1.TemplateStatus{},
			},
			seedWorkflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{},
			},
			seedHardware: &v1alpha1.Hardware{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Hardware",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine1",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							Netboot: &v1alpha1.Netboot{
								AllowPXE:      &[]bool{true}[0],
								AllowWorkflow: &[]bool{true}[0],
							},
							DHCP: &v1alpha1.DHCP{
								Arch:     "x86_64",
								Hostname: "sm01",
								IP: &v1alpha1.IP{
									Address: "172.16.10.100",
									Gateway: "172.16.10.1",
									Netmask: "255.255.255.0",
								},
								LeaseTime:   86400,
								MAC:         "3c:ec:ef:4c:4f:54",
								NameServers: []string{},
								UEFI:        true,
							},
						},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "debian",
					Namespace: "default",
				},
			},
			want: reconcile.Result{},
			wantWflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1000",
					Name:            "debian",
					Namespace:       "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         v1alpha1.WorkflowStatePending,
					GlobalTimeout: 1800,
					Tasks: []v1alpha1.Task{
						{
							Name: "os-installation",

							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Volumes: []string{
								"/dev:/dev",
								"/dev/console:/dev/console",
								"/lib/firmware:/lib/firmware:ro",
							},
							Actions: []v1alpha1.Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 600,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status: v1alpha1.WorkflowStatePending,
								},
							},
						},
					},
				},
			},
			wantErr: errors.New("parsing yaml data: yaml: line 2: found character that cannot start any token"),
		},
		{
			name: "MissingTemplate",
			seedTemplate: &v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "default",
				},
				Spec:   v1alpha1.TemplateSpec{},
				Status: v1alpha1.TemplateStatus{},
			},
			seedWorkflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian", // doesn't exist
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{},
			},
			seedHardware: &v1alpha1.Hardware{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Hardware",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine1",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							Netboot: &v1alpha1.Netboot{
								AllowPXE:      &[]bool{true}[0],
								AllowWorkflow: &[]bool{true}[0],
							},
							DHCP: &v1alpha1.DHCP{
								Arch:     "x86_64",
								Hostname: "sm01",
								IP: &v1alpha1.IP{
									Address: "172.16.10.100",
									Gateway: "172.16.10.1",
									Netmask: "255.255.255.0",
								},
								LeaseTime:   86400,
								MAC:         "3c:ec:ef:4c:4f:54",
								NameServers: []string{},
								UEFI:        true,
							},
						},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "debian",
					Namespace: "default",
				},
			},
			want: reconcile.Result{},
			wantWflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "999",
				},
			},
			wantErr: errors.New("no template found: name=debian; namespace=default"),
		},
		{
			name: "TimedOutWorkflow",
			seedTemplate: &v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.TemplateSpec{
					Data: &minimalTemplate,
				},
				Status: v1alpha1.TemplateStatus{},
			},
			seedWorkflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         v1alpha1.WorkflowStateRunning,
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "os-installation",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Volumes: []string{
								"/dev:/dev",
								"/dev/console:/dev/console",
								"/lib/firmware:/lib/firmware:ro",
							},
							Actions: []v1alpha1.Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    v1alpha1.WorkflowStateRunning,
									StartedAt: TestTime.MetaV1BeforeSec(601),
								},
							},
						},
					},
				},
			},
			seedHardware: &v1alpha1.Hardware{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Hardware",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine1",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							Netboot: &v1alpha1.Netboot{
								AllowPXE:      &[]bool{true}[0],
								AllowWorkflow: &[]bool{true}[0],
							},
							DHCP: &v1alpha1.DHCP{
								Arch:     "x86_64",
								Hostname: "sm01",
								IP: &v1alpha1.IP{
									Address: "172.16.10.100",
									Gateway: "172.16.10.1",
									Netmask: "255.255.255.0",
								},
								LeaseTime:   86400,
								MAC:         "3c:ec:ef:4c:4f:54",
								NameServers: []string{},
								UEFI:        true,
							},
						},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "debian",
					Namespace: "default",
				},
			},
			want: reconcile.Result{},
			wantWflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1000",
					Name:            "debian",
					Namespace:       "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         v1alpha1.WorkflowStateTimeout,
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "os-installation",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Volumes: []string{
								"/dev:/dev",
								"/dev/console:/dev/console",
								"/lib/firmware:/lib/firmware:ro",
							},
							Actions: []v1alpha1.Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 60,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status:    v1alpha1.WorkflowStateTimeout,
									StartedAt: TestTime.MetaV1BeforeSec(601),
									Seconds:   601,
									Message:   "Action timed out",
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ErrorGettingHardwareRef",
			seedTemplate: &v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.TemplateSpec{
					Data: &minimalTemplate,
				},
				Status: v1alpha1.TemplateStatus{},
			},
			seedWorkflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareRef: "i_dont_exist",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "debian",
					Namespace: "default",
				},
			},
			want: reconcile.Result{},
			wantWflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1000",
					Name:            "debian",
					Namespace:       "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         v1alpha1.WorkflowStatePending,
					GlobalTimeout: 1800,
					Tasks: []v1alpha1.Task{
						{
							Name: "os-installation",

							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Volumes: []string{
								"/dev:/dev",
								"/dev/console:/dev/console",
								"/lib/firmware:/lib/firmware:ro",
							},
							Actions: []v1alpha1.Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 600,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status: v1alpha1.WorkflowStatePending,
								},
							},
						},
					},
				},
			},
			wantErr: errors.New("hardware not found: name=i_dont_exist; namespace=default"),
		},
		{
			name: "SuccessWithHardwareRef",
			seedHardware: &v1alpha1.Hardware{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Hardware",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine1",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Disks: []v1alpha1.Disk{
						{Device: "/dev/nvme0n1"},
					},
					Interfaces: []v1alpha1.Interface{
						{
							Netboot: &v1alpha1.Netboot{
								AllowPXE:      &[]bool{true}[0],
								AllowWorkflow: &[]bool{true}[0],
							},
							DHCP: &v1alpha1.DHCP{
								Arch:     "x86_64",
								Hostname: "sm01",
								IP: &v1alpha1.IP{
									Address: "172.16.10.100",
									Gateway: "172.16.10.1",
									Netmask: "255.255.255.0",
								},
								LeaseTime:   86400,
								MAC:         "3c:ec:ef:4c:4f:54",
								NameServers: []string{},
								UEFI:        true,
							},
						},
					},
				},
			},
			seedTemplate: &v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.TemplateSpec{
					Data: &templateWithDiskTemplate,
				},
				Status: v1alpha1.TemplateStatus{},
			},
			seedWorkflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareRef: "machine1",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "debian",
					Namespace: "default",
				},
			},
			want: reconcile.Result{},
			wantWflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1000",
					Name:            "debian",
					Namespace:       "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareRef: "machine1",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         v1alpha1.WorkflowStatePending,
					GlobalTimeout: 1800,
					Tasks: []v1alpha1.Task{
						{
							Name: "os-installation",

							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Volumes: []string{
								"/dev:/dev",
								"/dev/console:/dev/console",
								"/lib/firmware:/lib/firmware:ro",
							},
							Actions: []v1alpha1.Action{
								{
									Name:    "stream-debian-image",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 600,
									Environment: map[string]string{
										"COMPRESSED": "true",
										"DEST_DISK":  "/dev/nvme0n1",
										"IMG_URL":    "http://10.1.1.11:8080/debian-10-openstack-amd64.raw.gz",
									},
									Status: v1alpha1.WorkflowStatePending,
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
		kc := GetFakeClientBuilder()
		if tc.seedHardware != nil {
			kc = kc.WithObjects(tc.seedHardware)
		}
		if tc.seedTemplate != nil {
			kc = kc.WithObjects(tc.seedTemplate)
		}
		if tc.seedWorkflow != nil {
			kc = kc.WithObjects(tc.seedWorkflow)
			kc = kc.WithStatusSubresource(tc.seedWorkflow)
		}
		controller := &Reconciler{
			client:  kc.Build(),
			nowFunc: TestTime.Now,
		}

		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := controller.Reconcile(context.Background(), tc.req)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}
			if tc.want != got {
				t.Errorf("Got unexpected result. Wanted %v, got %v", tc.want, got)
				// Don't return, also check the modified object
			}
			wflow := &v1alpha1.Workflow{}
			err := controller.client.Get(
				context.Background(),
				client.ObjectKey{Name: tc.wantWflow.Name, Namespace: tc.wantWflow.Namespace},
				wflow)
			if err != nil {
				t.Errorf("Error finding desired workflow: %v", err)
				return
			}

			if diff := cmp.Diff(tc.wantWflow, wflow); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}
