package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}

type (
	WorkflowState         string
	WorkflowConditionType string
)

const (
	WorkflowStatePending   = WorkflowState("STATE_PENDING")
	WorkflowStateRunning   = WorkflowState("STATE_RUNNING")
	WorkflowStateFailed    = WorkflowState("STATE_FAILED")
	WorkflowStateTimeout   = WorkflowState("STATE_TIMEOUT")
	WorkflowStateSuccess   = WorkflowState("STATE_SUCCESS")
	WorkflowStatePreparing = WorkflowState("STATE_PREPARING")

	NetbootJobFailed        WorkflowConditionType = "NetbootJobFailed"
	NetbootJobComplete      WorkflowConditionType = "NetbootJobComplete"
	NetbootJobRunning       WorkflowConditionType = "NetbootJobRunning"
	NetbootJobSetupFailed   WorkflowConditionType = "NetbootJobSetupFailed"
	NetbootJobSetupComplete WorkflowConditionType = "NetbootJobSetupComplete"
	ToggleAllowNetbootTrue  WorkflowConditionType = "AllowNetbootTrue"
	ToggleAllowNetbootFalse WorkflowConditionType = "AllowNetbootFalse"
	TemplateRenderedSuccess WorkflowConditionType = "TemplateRenderedSuccess"
)

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=workflows,scope=Namespaced,categories=tinkerbell,shortName=wf,singular=workflow
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:JSONPath=".spec.templateRef",name=Template,type=string
// +kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string
// +kubebuilder:printcolumn:JSONPath=".status.currentAction",name=Current-Action,type=string
// +kubebuilder:printcolumn:JSONPath=".status.templateRending",name=Template-Rendering,type=string

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

// WorkflowSpec defines the desired state of Workflow.
type WorkflowSpec struct {
	// Name of the Template associated with this workflow.
	TemplateRef string `json:"templateRef,omitempty"`

	// Name of the Hardware associated with this workflow.
	// +optional
	HardwareRef string `json:"hardwareRef,omitempty"`

	// A mapping of template devices to hadware mac addresses
	HardwareMap map[string]string `json:"hardwareMap,omitempty"`

	// BootOpts is a set of options to be used when netbooting the hardware.
	BootOpts BootOpts `json:"bootOpts,omitempty"`
}

type BootOpts struct {
	// ToggleAllowNetboot indicates whether the controller should toggle the field in the associated hardware for allowing PXE booting.
	// This will be enabled before a Workflow is executed and disabled after the Workflow has completed successfully.
	// A HardwareRef must be provided.
	// +optional
	ToggleAllowNetboot bool `json:"toggleAllowNetboot,omitempty"`
	// OneTimeNetboot indicates whether the controller should create a job.bmc.tinkerbell.org object for getting the associated hardware
	// into a netbooting state.
	// A HardwareRef that contains a spec.BmcRef must be provided.
	// +optional
	OneTimeNetboot bool `json:"oneTimeNetboot,omitempty"`
}

// WorkflowStatus defines the observed state of Workflow.
type WorkflowStatus struct {
	// State is the state of the workflow in Tinkerbell.
	State WorkflowState `json:"state,omitempty"`

	// CurrentAction is the action that is currently in the running state.
	CurrentAction string `json:"currentAction,omitempty"`

	// JobUID is the UID of the BMCJob associated with this workflow.
	// This is used to identify the unique job.bmc.tinkerbell.org object, as
	// all objects are created with the same name.
	// JobUID types.UID `json:"jobUid,omitempty"`

	// JobComplete        bool `json:"jobComplete,omitempty"`
	// ExistingJobDeleted bool `json:"existingJobDeleted,omitempty"`

	Job JobStatus `json:"jobStatus,omitempty"`

	// TemplateRendering indicates whether the template was rendered successfully.
	// Possible values are "successful" or "failed".
	TemplateRendering string `json:"templateRending,omitempty"`

	// GlobalTimeout represents the max execution time
	GlobalTimeout int64 `json:"globalTimeout,omitempty"`

	// Tasks are the tasks to be completed
	Tasks []Task `json:"tasks,omitempty"`

	// The latest available observations of an object's current state. When a Job
	// fails, one of the conditions will have type "Failed" and status true. When
	// a Job is suspended, one of the conditions will have type "Suspended" and
	// status true; when the Job is resumed, the status of this condition will
	// become false. When a Job is completed, one of the conditions will have
	// type "Complete" and status true.
	//
	// A job is considered finished when it is in a terminal condition, either
	// "Complete" or "Failed". A Job cannot have both the "Complete" and "Failed" conditions.
	// Additionally, it cannot be in the "Complete" and "FailureTarget" conditions.
	// The "Complete", "Failed" and "FailureTarget" conditions cannot be disabled.
	//
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=atomic
	Conditions []WorkflowCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

type JobStatus struct {
	// UID is the UID of the BMCJob associated with this workflow.
	// This is used to identify the unique job.bmc.tinkerbell.org object, as
	// all objects are created with the same name.
	UID types.UID `json:"uid,omitempty"`

	// Complete indicates whether the created job.bmc.tinkerbell.org has reported its conditions as complete.
	Complete bool `json:"complete,omitempty"`

	// ExistingJobDeleted indicates whether any existing job.bmc.tinkerbell.org was deleted.
	// The name of each job.bmc.tinkerbell.org object is the same, so only one can exist at a time.
	// Using the same name was chosen so that there is only ever 1 job.bmc.tinkerbell.org per hardware.
	// This makes clean up easier and we dont just orphan jobs every time.
	ExistingJobDeleted bool `json:"existingJobDeleted,omitempty"`
}

// JobCondition describes current state of a job.
type WorkflowCondition struct {
	// Type of job condition, Complete or Failed.
	Type WorkflowConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=WorkflowConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status metav1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
	// Time when the condition was created.
	// +optional
	Time *metav1.Time `json:"time,omitempty" protobuf:"bytes,7,opt,name=time"`
}

// Wanted to use metav1.Status but kubebuilder errors with, "must apply listType to an array, found".
type Status struct {
	// Status of the operation.
	// One of: "Success" or "Failure".
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status string `json:"status,omitempty" protobuf:"bytes,2,opt,name=status"`
	// A human-readable description of the status of this operation.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
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

// HasCondition checks if the cType condition is present with status cStatus on a bmj.
func (w *WorkflowStatus) HasCondition(wct WorkflowConditionType, cs metav1.ConditionStatus) bool {
	for _, c := range w.Conditions {
		if c.Type == wct {
			return c.Status == cs
		}
	}

	return false
}

// SetCondition updates conditions. If the condition already exists, it updates it.
// If the condition doesn't exist then it appends the new one (wc).
func (w *WorkflowStatus) SetCondition(wc WorkflowCondition) {
	index := -1
	for i, c := range w.Conditions {
		if c.Type == wc.Type {
			index = i
			break
		}
	}
	if index != -1 {
		w.Conditions[index] = wc
		return
	}

	w.Conditions = append(w.Conditions, wc)
}
