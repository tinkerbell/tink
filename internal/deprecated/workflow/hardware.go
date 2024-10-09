package workflow

import (
	"context"
	"fmt"

	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// setAllowPXE sets the allowPXE field on the hardware network interfaces.
// If hardware is nil then it will be retrieved using the client.
func setAllowPXE(ctx context.Context, cc client.Client, w *v1alpha1.Workflow, h *v1alpha1.Hardware, allowPXE bool) error {
	if h == nil && w == nil {
		return fmt.Errorf("both workflow and hardware cannot be nil")
	}
	if h == nil {
		h = &v1alpha1.Hardware{}
		if err := cc.Get(ctx, client.ObjectKey{Name: w.Spec.HardwareRef, Namespace: w.Namespace}, h); err != nil {
			return fmt.Errorf("hardware not found: name=%v; namespace=%v, error: %w", w.Spec.HardwareRef, w.Namespace, err)
		}
	}

	for _, iface := range h.Spec.Interfaces {
		iface.Netboot.AllowPXE = ptr.Bool(allowPXE)
	}

	if err := cc.Update(ctx, h); err != nil {
		return fmt.Errorf("error updating allow pxe: %w", err)
	}

	return nil
}

// hardwareFrom retrieves the in cluster hardware object defined in the given workflow.
func hardwareFrom(ctx context.Context, cc client.Client, w *v1alpha1.Workflow) (*v1alpha1.Hardware, error) {
	if w == nil {
		return nil, fmt.Errorf("workflow is nil")
	}
	if w.Spec.HardwareRef == "" {
		return nil, fmt.Errorf("hardware ref is empty")
	}
	h := &v1alpha1.Hardware{}
	if err := cc.Get(ctx, client.ObjectKey{Name: w.Spec.HardwareRef, Namespace: w.Namespace}, h); err != nil {
		return nil, fmt.Errorf("hardware not found: name=%v; namespace=%v, error: %w", w.Spec.HardwareRef, w.Namespace, err)
	}

	return h, nil
}

// toggleHardware toggles the allowPXE field on the hardware network interfaces.
// It is idempotent and uses the Workflow.Status.BootOptionsStatus.AllowNetboot fields for idempotent checks.
func (s *state) toggleHardware(ctx context.Context, allowPXE bool) (v1alpha1.WorkflowCondition, error) {
	// 1. check if we've already set the allowPXE field to the desired value
	// 2. if not, set the allowPXE field to the desired value
	// 3. return a WorkflowCondition with the result of the operation
	var condType v1alpha1.WorkflowConditionType
	switch allowPXE {
	case true:
		condType = v1alpha1.ToggleAllowNetbootTrue
		if s.workflow.Status.BootOptions.AllowNetboot.ToggledTrue {
			return v1alpha1.WorkflowCondition{
				Type:    condType,
				Status:  metav1.ConditionTrue,
				Reason:  "Complete",
				Message: "allowPXE already set to true",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			}, nil
		}
	case false:
		condType = v1alpha1.ToggleAllowNetbootFalse
		if s.workflow.Status.BootOptions.AllowNetboot.ToggledFalse {
			return v1alpha1.WorkflowCondition{
				Type:    condType,
				Status:  metav1.ConditionTrue,
				Reason:  "Complete",
				Message: "allowPXE already set to false",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			}, nil
		}
	}
	if err := setAllowPXE(ctx, s.client, s.workflow, s.hardware, allowPXE); err != nil {
		return v1alpha1.WorkflowCondition{
			Type:    condType,
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: fmt.Sprintf("error setting allowPXE to %v: %v", allowPXE, err),
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		}, err
	}
	return v1alpha1.WorkflowCondition{
		Type:    condType,
		Status:  metav1.ConditionTrue,
		Reason:  "Complete",
		Message: fmt.Sprintf("set allowPXE to %v", allowPXE),
		Time:    &metav1.Time{Time: metav1.Now().UTC()},
	}, nil
}
