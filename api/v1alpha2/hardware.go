package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HardwareSpec struct {
	// NetworkInterfaces defines the desired DHCP and netboot configuration for a network interface.
	// +kubebuilder:validation:MinPoperties=1
	NetworkInterfaces NetworkInterfaces `json:"networkInterfaces,omitempty"`

	// IPXE provides iPXE script override fields. This is useful for debugging or netboot
	// customization.
	// +optional.
	IPXE *IPXE `json:"ipxe,omitempty"`

	// OSIE describes the Operating System Installation Environment to be netbooted.
	OSIE corev1.LocalObjectReference `json:"osie,omitempty"`

	// KernelParams passed to the kernel when launching the OSIE. Parameters are joined with a
	// space.
	// +optional
	KernelParams []string `json:"kernelParams,omitempty"`

	// Instance describes instance specific data that is generally unused by Tinkerbell core.
	// +optional
	Instance *Instance `json:"instance,omitempty"`

	// StorageDevices is a list of storage devices that will be available in the OSIE.
	// +optional.
	StorageDevices []StorageDevice `json:"storageDevices,omitempty"`

	// BMCRef references a Rufio Machine object.
	// +optional.
	BMCRef *corev1.LocalObjectReference `json:"bmcRef,omitempty"`
}

// NetworkInterfaces maps a MAC address to a NetworkInterface.
type NetworkInterfaces map[MAC]NetworkInterface

// NetworkInterface is the desired configuration for a particular network interface.
type NetworkInterface struct {
	// DHCP is the basic network information for serving DHCP requests. Required when DisbaleDHCP
	// is false.
	// +optional
	DHCP *DHCP `json:"dhcp,omitempty"`

	// DisableDHCP disables DHCP for this interface. Implies DisableNetboot.
	// +kubebuilder:default=false
	DisableDHCP bool `json:"disableDhcp,omitempty"`

	// DisableNetboot disables netbooting for this interface. The interface will still receive
	// network information specified by DHCP.
	// +kubebuilder:default=false
	DisableNetboot bool `json:"disableNetboot,omitempty"`
}

// IsDHCPEnabled checks if DHCP is enabled for ni.
func (ni NetworkInterface) IsDHCPEnabled() bool {
	return !ni.DisableDHCP
}

// IsNetbootEnabled checks if Netboot is enabled for ni.
func (ni NetworkInterface) IsNetbootEnabled() bool {
	return !ni.DisableNetboot
}

// DHCP describes basic network configuration to be served in DHCP OFFER responses. It can be
// considered a DHCP reservation.
type DHCP struct {
	// IP is an IPv4 address to serve.
	// +kubebuilder:validation:Pattern=`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`
	IP string `json:"ip,omitempty"`

	// Netmask is an IPv4 netmask to serve.
	// +kubebuilder+validation:Pattern=`^(255)\.(0|128|192|224|240|248|252|254|255)\.(0|128|192|224|240|248|252|254|255)\.(0|128|192|224|240|248|252|254|255)`
	Netmask string `json:"netmask,omitempty"`

	// Gateway is the default gateway address to serve.
	// +kubebuilder:validation:Pattern=`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`
	// +optional
	Gateway *string `json:"gateway,omitempty"`

	// +kubebuilder:validation:Pattern=`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9]"[A-Za-z0-9\-]*[A-Za-z0-9])$`
	// +optional
	Hostname *string `json:"hostname,omitempty"`

	// VLANID is a VLAN ID between 0 and 4096.
	// +kubebuilder:validation:Pattern=`^(([0-9][0-9]{0,2}|[1-3][0-9][0-9][0-9]|40([0-8][0-9]|9[0-6]))(,[1-9][0-9]{0,2}|[1-3][0-9][0-9][0-9]|40([0-8][0-9]|9[0-6]))*)$`
	// +optional
	VLANID *string `json:"vlanId,omitempty"`

	// Nameservers to serve.
	// +optional
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// Timeservers to serve.
	// +optional
	Timeservers []Timeserver `json:"timeservers,omitempty"`

	// LeaseTimeSeconds to serve. 24h default. Maximum equates to max uint32 as defined by RFC 2132
	// § 9.2 (https://www.rfc-editor.org/rfc/rfc2132.html#section-9.2).
	// +kubebuilder:default=86400
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +optional
	LeaseTimeSeconds *int64 `json:"leaseTimeSeconds"`
}

// IPXE describes overrides for IPXE scripts. At least 1 option must be specified.
type IPXE struct {
	// Content is an inline iPXE script.
	// +optional
	Content *string `json:"inline,omitempty"`

	// URL is a URL to a hosted iPXE script.
	// +optional
	URL *string `json:"url,omitempty"`
}

// Instance describes instance specific data. Instance specific data is typically dependent on the
// permanent OS that a piece of hardware runs. This data is often served by an instance metadata
// service such as Tinkerbell's Hegel. The core Tinkerbell stack does not leverage this data.
type Instance struct {
	// Userdata is data with a structure understood by the producer and consumer of the data.
	// +optional
	Userdata *string `json:"userdata,omitempty"`

	// Vendordata is data with a structure understood by the producer and consumer of the data.
	// +optional
	Vendordata *string `json:"vendordata,omitempty"`
}

// MAC is a Media Access Control address. MACs must use lower case letters.
// +kubebuilder:validation:Pattern=`^([0-9a-f]{2}:){5}([0-9a-f]{2})$`
type MAC string

// Nameserver is an IP or hostname.
// +kubebuilder:validation:Pattern=`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$|^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
type Nameserver string

// Timeserver is an IP or hostname.
// +kubebuilder:validation:Pattern=`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$|^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
type Timeserver string

// StorageDevice describes a storage device path that will be present in the OSIE.
// StorageDevices must be valid Linux paths. They should not contain partitions.
//
// Good
//
//	/dev/sda
//	/dev/nvme0n1
//
// Bad (contains partitions)
//
//	/dev/sda1
//	/dev/nvme0n1p1
//
// Bad (invalid Linux path)
//
//	\dev\sda
//
// +kubebuilder:validation:Pattern=`^(/[^/ ]*)+/?$`
type StorageDevice string

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=tinkerbell,path=hardware,shortName=hw
// +kubebuilder:printcolumn:name="BMC",type="string",JSONPath=".spec.bmcRef",description="Baseboard management computer attached to the Hardware"
// +kubebuilder:unservedversion

// Hardware is a logical representation of a machine that can execute Workflows.
type Hardware struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HardwareSpec `json:"spec,omitempty"`
}

// GetMACs retrieves all MACs associated with h.
func (h *Hardware) GetMACs() []string {
	var macs []string
	for m := range h.Spec.NetworkInterfaces {
		macs = append(macs, string(m))
	}
	return macs
}

// GetIPs retrieves all IP addresses. It does not consider the DisableDHCP flag.
func (h *Hardware) GetIPs() []string {
	var ips []string
	for _, ni := range h.Spec.NetworkInterfaces {
		if ni.DHCP != nil {
			ips = append(ips, ni.DHCP.IP)
		}
	}
	return ips
}

// +kubebuilder:object:root=true

type HardwareList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hardware `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hardware{}, &HardwareList{})
}
