package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TemplateSpec struct {
	// Actions defines the set of actions to be run on a target machine. Actions are run sequentially
	// in the order they are specified. At least 1 action must be specified. Names of actions
	// must be unique within a Template.
	// +kubebuilder:validation:MinItems=1
	Actions []Action `json:"actions,omitempty"`

	// Volumes to be mounted on all actions. If an action specifies the same volume it will take
	// precedence.
	// +optional
	Volumes []Volume `json:"volumes,omitempty"`

	// Env defines environment variables to be available in all actions. If an action specifies
	// the same environment variable it will take precedence.
	// +optional
	Env map[string]string `json:"env,omitempty"`
}

// Action defines an individual action to be run on a target machine.
type Action struct {
	// Name is a name for the action.
	Name string `json:"name"`

	// Image is an OCI image.
	Image string `json:"image"`

	// Cmd defines the command to use when launching the image. It overrides the default command
	// of the action. It must be a unix path to an executable program.
	// +kubebuilder:validation:Pattern=`^(/[^/ ]*)+/?$`
	// +optional
	Cmd *string `json:"cmd,omitempty"`

	// Args are a set of arguments to be passed to the command executed by the container on
	// launch.
	// +optional
	Args []string `json:"args,omitempty"`

	// Env defines environment variables used when launching the container.
	//+optional
	Env map[string]string `json:"env,omitempty"`

	// Volumes defines the volumes to mount into the container.
	// +optional
	Volumes []Volume `json:"volumes,omitempty"`

	// Namespace defines the Linux namespaces this container should execute in.
	// +optional
	Namespace *Namespace `json:"namespaces,omitempty"`
}

// Volume is a specification for mounting a volume in an action. Volumes take the form
// {SRC-VOLUME-NAME | SRC-HOST-DIR}:TGT-CONTAINER-DIR:OPTIONS. When specifying a VOLUME-NAME that
// does not exist it will be created for you. Examples:
//
// Read-only bind mount bound to /data
//
//	/etc/data:/data:ro
//
// Writable volume name bound to /data
//
//	shared_volume:/data
//
// See https://docs.docker.com/storage/volumes/ for additional details.
type Volume string

// Namespace defines the Linux namespaces to use for the container.
// See https://man7.org/linux/man-pages/man7/namespaces.7.html.
type Namespace struct {
	// Network defines the network namespace.
	// +optional
	Network *string `json:"network,omitempty"`

	// PID defines the PID namespace
	// +optional
	PID *int `json:"pid,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=tinkerbell,shortName=tpl
// +kubebuilder:unservedversion

// Template defines a set of actions to be run on a target machine. The template is rendered
// prior to execution where it is exposed to Hardware and user defined data. Most fields within the
// TemplateSpec may contain templates values excluding .TemplateSpec.Actions[].Name.
// See https://pkg.go.dev/text/template for more details.
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
