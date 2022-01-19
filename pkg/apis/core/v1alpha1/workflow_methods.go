package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// WorkflowIDAnnotation is used by the controller to store the
	// ID assigned to the workflow by Tinkerbell for migrated workflows.
	WorkflowIDAnnotation = "workflow.tinkerbell.org/id"
)

// TinkID returns the Tinkerbell ID associated with this Workflow.
func (w *Workflow) TinkID() string {
	return w.Annotations[WorkflowIDAnnotation]
}

// SetTinkID sets the Tinkerbell ID associated with this Workflow.
func (w *Workflow) SetTinkID(id string) {
	if w.Annotations == nil {
		w.Annotations = make(map[string]string)
	}
	w.Annotations[WorkflowIDAnnotation] = id
}

// GetStartTime returns the start time, for the first action of the first task.
func (w *Workflow) GetStartTime() *metav1.Time {
	if len(w.Status.Tasks) > 0 {
		if len(w.Status.Tasks[0].Actions) > 0 {
			return w.Status.Tasks[0].Actions[0].StartedAt
		}
	}
	return nil
}
