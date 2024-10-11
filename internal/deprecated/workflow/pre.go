package workflow

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// prepareWorkflow prepares the workflow for execution.
// The workflow (s.workflow) can be updated even if an error occurs.
// Any patching of the workflow object in a cluster is left up to the caller.
func (s *state) prepareWorkflow(ctx context.Context) (reconcile.Result, error) {
	tracer := otel.Tracer("prepareWorkflow")
	var span trace.Span
	ctx, span = tracer.Start(ctx, "prepareWorkflow")
	defer span.End()
	// handle bootoptions
	// 1. Handle toggling allowPXE in a hardware object if toggleAllowNetboot is true.
	if s.workflow.Spec.BootOptions.ToggleAllowNetboot {
		s.logger.InfoContext(ctx, "toggling allowPXE true")
		if err := s.toggleHardware(ctx, true); err != nil {
			return reconcile.Result{}, err
		}
	}

	// 2. Handle booting scenarios.
	switch s.workflow.Spec.BootOptions.BootMode {
	case v1alpha1.BootModeNetboot:
		s.logger.InfoContext(ctx, "boot mode netboot")
		if s.hardware == nil {
			return reconcile.Result{}, errors.New("hardware is nil")
		}
		name := jobName(fmt.Sprintf("%s-%s", jobNameNetboot, s.hardware.Name))
		efiBoot := func() bool {
			for _, iface := range s.hardware.Spec.Interfaces {
				if iface.DHCP != nil && iface.DHCP.UEFI {
					return true
				}
			}
			return false
		}()
		actions := []rufio.Action{
			{
				PowerAction: rufio.PowerHardOff.Ptr(),
			},
			{
				OneTimeBootDeviceAction: &rufio.OneTimeBootDeviceAction{
					Devices: []rufio.BootDevice{
						rufio.PXE,
					},
					EFIBoot: efiBoot,
				},
			},
			{
				PowerAction: rufio.PowerOn.Ptr(),
			},
		}

		r, err := s.handleJob(ctx, actions, name)
		if s.workflow.Status.BootOptions.Jobs[name.String()].Complete && s.workflow.Status.State == v1alpha1.WorkflowStatePreparing {
			s.workflow.Status.State = v1alpha1.WorkflowStatePending
		}
		return r, err
	case v1alpha1.BootModeISO:
		s.logger.InfoContext(ctx, "boot mode iso")
		if s.hardware == nil {
			return reconcile.Result{}, errors.New("hardware is nil")
		}
		if s.workflow.Spec.BootOptions.ISOURL == "" {
			return reconcile.Result{}, errors.New("iso url must be a valid url")
		}
		name := jobName(fmt.Sprintf("%s-%s", jobNameISOMount, s.hardware.Name))
		efiBoot := func() bool {
			for _, iface := range s.hardware.Spec.Interfaces {
				if iface.DHCP != nil && iface.DHCP.UEFI {
					return true
				}
			}
			return false
		}()
		actions := []rufio.Action{
			{
				PowerAction: rufio.PowerHardOff.Ptr(),
			},
			{
				VirtualMediaAction: &rufio.VirtualMediaAction{
					MediaURL: "", // empty to unmount/eject the media
					Kind:     rufio.VirtualMediaCD,
				},
			},
			{
				VirtualMediaAction: &rufio.VirtualMediaAction{
					MediaURL: s.workflow.Spec.BootOptions.ISOURL,
					Kind:     rufio.VirtualMediaCD,
				},
			},
			{
				OneTimeBootDeviceAction: &rufio.OneTimeBootDeviceAction{
					Devices: []rufio.BootDevice{
						rufio.CDROM,
					},
					EFIBoot: efiBoot,
				},
			},
			{
				PowerAction: rufio.PowerOn.Ptr(),
			},
		}

		r, err := s.handleJob(ctx, actions, name)
		if s.workflow.Status.BootOptions.Jobs[name.String()].Complete && s.workflow.Status.State == v1alpha1.WorkflowStatePreparing {
			s.workflow.Status.State = v1alpha1.WorkflowStatePending
		}
		return r, err
	}

	return reconcile.Result{}, nil
}
