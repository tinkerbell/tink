package hardware

import (
	"fmt"
	"net/http"

	"github.com/tinkerbell/tink/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (a *Admission) validateConditionalFields(hw *v1alpha2.Hardware) admission.Response {
	for mac, ni := range hw.Spec.NetworkInterfaces {
		if ni.IsDHCPEnabled() && ni.DHCP == nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf(
				"network interface for %v has DHCP enabled but no DHCP config",
				mac,
			))
		}
	}

	return admission.Allowed("")
}
