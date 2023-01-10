package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkflowState string

const (
	WorkflowStatePending = WorkflowState("STATE_PENDING")
	WorkflowStateRunning = WorkflowState("STATE_RUNNING")
	WorkflowStateFailed  = WorkflowState("STATE_FAILED")
	WorkflowStateTimeout = WorkflowState("STATE_TIMEOUT")
	WorkflowStateSuccess = WorkflowState("STATE_SUCCESS")
)

// WorkflowSpec defines the desired state of Workflow.
type WorkflowSpec struct {
	// Name of the Template associated with this workflow.
	TemplateRef string `json:"templateRef,omitempty"`

	// Name of the Hardware associated with this workflow.
	HardwareRef string `json:"hardwareRef,omitempty"`

	// A mapping of template devices to hadware mac addresses
	HardwareMap map[string]string `json:"hardwareMap,omitempty"`
}

// WorkflowStatus defines the observed state of Workflow.
type WorkflowStatus struct {
	// State is the state of the workflow in Tinkerbell.
	State WorkflowState `json:"state,omitempty"`

	// GlobalTimeout represents the max execution time
	GlobalTimeout int64 `json:"globalTimeout,omitempty"`

	// Tasks are the tasks to be completed
	Tasks []Task `json:"tasks,omitempty"`
}

// Task represents a series of actions to be completed by a worker.
type Task struct {
	Name        string            `json:"name"`
	WorkerAddr  string            `json:"worker"`
	Actions     []Action          `json:"actions"`
	Volumes     []string          `json:"volumes,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// Action represents a workflow action.
type Action struct {
	Name        string            `json:"name,omitempty"`
	Image       string            `json:"image,omitempty"`
	Timeout     int64             `json:"timeout,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Volumes     []string          `json:"volumes,omitempty"`
	Pid         string            `json:"pid,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Status      WorkflowState     `json:"status,omitempty"`
	StartedAt   *metav1.Time      `json:"startedAt,omitempty"`
	Seconds     int64             `json:"seconds,omitempty"`
	Message     string            `json:"message,omitempty"`
}

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=workflows,scope=Namespaced,categories=tinkerbell,shortName=wf,singular=workflow
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:JSONPath=".spec.templateRef",name=Template,type=string
// +kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string

// Workflow is the Schema for the Workflows API.
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflows.
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}
