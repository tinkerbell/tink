package workflow

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (s *state) postActions(ctx context.Context) (reconcile.Result, error) {
	// 1. Handle toggling allowPXE in a hardware object if toggleAllowNetboot is true.
	if s.workflow.Spec.BootOptions.ToggleAllowNetboot {
		if err := s.toggleHardware(ctx, false); err != nil {
			return reconcile.Result{}, err
		}
	}

	// 2. Handle ISO eject scenario.
	switch s.workflow.Spec.BootOptions.BootMode {
	case v1alpha1.BootModeISO:
		if s.hardware == nil {
			return reconcile.Result{}, errors.New("hardware is nil")
		}
		if s.workflow.Spec.BootOptions.ISOURL == "" {
			return reconcile.Result{}, errors.New("iso url must be a valid url")
		}
		name := jobName(fmt.Sprintf("%s-%s", jobNameISOEject, s.hardware.Name))
		actions := []rufio.Action{
			{
				VirtualMediaAction: &rufio.VirtualMediaAction{
					MediaURL: "", // empty to unmount/eject the media
					Kind:     rufio.VirtualMediaCD,
				},
			},
		}

		r, err := s.handleJob(ctx, actions, name)
		if s.workflow.Status.BootOptions.Jobs[name.String()].Complete {
			s.workflow.Status.State = v1alpha1.WorkflowStateSuccess
		}
		return r, err
	}

	s.workflow.Status.State = v1alpha1.WorkflowStateSuccess
	return reconcile.Result{}, nil
}
