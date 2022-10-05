package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkflowSpec defines the desired state of Workflow.
type WorkflowDataSpec struct {
	// Name of the Workflow associated with this workflow.
	WorkflowRef string `json:"workflowRef,omitempty"`
}

// WorkflowStatus defines the observed state of Workflow.
type WorkflowDataStatus struct {
	// Data is the populated Workflow Data in Tinkerbell.
	Data string `json:"data,omitempty"`

	// Metadata is the metadata stored in Tinkerbell.
	Metadata string `json:"metadata,omitempty"`
}

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=workflowdata,scope=Namespaced,categories=tinkerbell,shortName=wfdata
// +kubebuilder:storageversion

// Workflow is the Schema for the Workflows API.
type WorkflowData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflows.
type WorkflowDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkflowData `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkflowData{}, &WorkflowDataList{})
}
