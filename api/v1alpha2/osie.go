package v1alpha2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type OSIESpec struct {
	// KernelURL is a URL to a kernel image.
	KernelURL string `json:"kernelUrl,omitempty"`

	// InitrdURL is a URL to an initrd image.
	InitrdURL string `json:"initrdUrl,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=tinkerbell

// OSIE describes an Operating System Installation Environment. It is used by Tinkerbell
// to provision machines and should launch the Tink Worker component.
type OSIE struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec OSIESpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type OSIEList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OSIE `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OSIE{}, &OSIEList{})
}
