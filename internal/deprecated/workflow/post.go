package workflow

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/deprecated/workflow/journal"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (s *state) postActions(ctx context.Context) (reconcile.Result, error) {
	// 1. Handle toggling allowPXE in a hardware object if toggleAllowNetboot is true.
	if s.workflow.Spec.BootOptions.ToggleAllowNetboot && !s.workflow.Status.BootOptions.AllowNetboot.ToggledFalse {
		journal.Log(ctx, "toggling allowPXE false")
		if err := s.toggleHardware(ctx, false); err != nil {
			return reconcile.Result{}, err
		}
	}

	// 2. Handle ISO eject scenario.
	if s.workflow.Spec.BootOptions.BootMode == v1alpha1.BootModeISO {
		name := jobName(fmt.Sprintf("%s-%s", jobNameISOEject, s.workflow.GetName()))
		if j := s.workflow.Status.BootOptions.Jobs[name.String()]; !j.ExistingJobDeleted || j.UID == "" || !j.Complete {
			journal.Log(ctx, "boot mode iso")
			if s.workflow.Spec.BootOptions.ISOURL == "" {
				return reconcile.Result{}, errors.New("iso url must be a valid url")
			}
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
	}

	s.workflow.Status.State = v1alpha1.WorkflowStateSuccess
	return reconcile.Result{}, nil
}
