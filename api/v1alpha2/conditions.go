package v1alpha2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ConditionType identifies the type of condition.
type ConditionType string

// ConditionStatus expresses the current state of the condition.
type ConditionStatus string

const (
	// ConditionStatusUnknown is the default status and indicates the condition cannot be
	// evaluated as True or False.
	ConditionStatusUnknown ConditionStatus = "Unknown"

	// ConditionStatusTrue indicates the condition has been evaluated as true.
	ConditionStatusTrue ConditionStatus = "True"

	// ConditionStatusFalse indicates the condition has been evaluated as false.
	ConditionStatusFalse ConditionStatus = "False"
)

// Condition defines an observation on a resource that is generally attainable by inspecting
// other status fields.
type Condition struct {
	// Type of condition.
	Type ConditionType `json:"type"`

	// Status of the condition.
	Status ConditionStatus `json:"status"`

	// LastTransition is the last time the condition transitioned from one status to another.
	LastTransition *metav1.Time `json:"lastTransitionTime"`

	// Reason is a short CamelCase description for the conditions last transition.
	// +optional
	Reason *string `json:"reason,omitempty"`

	// Message is a human readable message indicating details about the last transition.
	// +optional
	Message *string `json:"message,omitempty"`
}

// Conditions define a list of observations of a particular resource.
type Conditions []Condition
