package server

import (
	"context"
	"strings"

	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type AutoCapMode string

var (
	AutoCapModeDiscovery  AutoCapMode = "discovery"
	AutoCapModeEnrollment AutoCapMode = "enrollment"
	AutoCapModeDisabled   AutoCapMode = "disabled"
)

func (k *KubernetesBackedServer) hardwareObjectExists(ctx context.Context, workerID string) bool {
	if err := k.ClientFunc().Get(ctx, types.NamespacedName{Name: strings.ReplaceAll(workerID, ":", "."), Namespace: k.namespace}, &v1alpha1.Hardware{}); err != nil {
		return false
	}
	return true
}

func (k *KubernetesBackedServer) createHardwareObject(ctx context.Context, workerID string) error {
	hw := &v1alpha1.Hardware{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ReplaceAll(workerID, ":", "."),
			Namespace: k.namespace,
		},
		Spec: v1alpha1.HardwareSpec{
			Interfaces: []v1alpha1.Interface{
				{
					DHCP: &v1alpha1.DHCP{
						MAC: workerID,
					},
					Netboot: &v1alpha1.Netboot{
						AllowPXE: ptr.Bool(true),
					},
				},
			},
		},
	}
	return k.ClientFunc().Create(ctx, hw)
}

func (k *KubernetesBackedServer) createWorkflowObject(ctx context.Context, workerID string) error {
	wf := &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ReplaceAll(workerID, ":", "."),
			Namespace: k.namespace,
		},
		Spec: v1alpha1.WorkflowSpec{
			HardwareRef: strings.ReplaceAll(workerID, ":", "."),
			TemplateRef: k.AutoEnrollmentTemplate,
			HardwareMap: map[string]string{
				"device_1": workerID,
			},
		},
	}
	return k.ClientFunc().Create(ctx, wf)
}
