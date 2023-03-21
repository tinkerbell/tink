package internal

import (
	"github.com/tinkerbell/tink/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HardwareByMACAddr is an index used with a controller-runtime client to lookup hardware by MAC.
const HardwareByMACAddr = ".Spec.NetworkInterfaces.MAC"

// HardwareByMACAddrFunc returns a list of MAC addresses for a Hardware object.
func HardwareByMACAddrFunc(obj client.Object) []string {
	hw, ok := obj.(*v1alpha2.Hardware)
	if !ok {
		return nil
	}
	return hw.GetMACs()
}

// HardwareByIPAddr is an index used with a controller-runtime client to lookup hardware by IP.
const HardwareByIPAddr = ".Spec.NetworkInterfaces.DHCP.IP"

// HardwareByIPAddrFunc returns a list of IP addresses for a Hardware object.
func HardwareByIPAddrFunc(obj client.Object) []string {
	hw, ok := obj.(*v1alpha2.Hardware)
	if !ok {
		return nil
	}
	return hw.GetIPs()
}
