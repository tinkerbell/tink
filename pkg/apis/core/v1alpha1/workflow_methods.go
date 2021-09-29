package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// GetStartTime returns the start time, for the first action of the first task.
func (w *Workflow) GetStartTime() *metav1.Time {
	if len(w.Status.Tasks) > 0 {
		if len(w.Status.Tasks[0].Actions) > 0 {
			return w.Status.Tasks[0].Actions[0].StartedAt
		}
	}
	return nil
}
