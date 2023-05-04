package hardware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/tinkerbell/tink/api/v1alpha2"
	"github.com/tinkerbell/tink/internal/hardware/internal"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (a *Admission) validateMACs(hw *v1alpha2.Hardware) admission.Response {
	// Validate all MACs on hw are valid before we compare them with Hardware in the cluster.
	if invalidMACs := getInvalidMACs(hw); len(invalidMACs) > 0 {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf(
			"invalid MAC address (%v): %v",
			macRegex.String(),
			strings.Join(invalidMACs, ", "),
		))
	}

	return admission.Allowed("")
}

// macRegex is taken from the API package documentation. It checks for valid MAC addresses.
// It expects MACs to be lowercase which is necessary for index lookups on API objects.
var macRegex = regexp.MustCompile("^([0-9a-f]{2}:){5}([0-9a-f]{2})$")

func getInvalidMACs(hw *v1alpha2.Hardware) []string {
	var invalidMACs []string
	for _, mac := range hw.GetMACs() {
		if mac == "" {
			mac = "<empty string>"
		}
		if !macRegex.MatchString(mac) {
			invalidMACs = append(invalidMACs, mac)
		}
	}
	return invalidMACs
}

func (a *Admission) validateUniqueMACs(ctx context.Context, hw *v1alpha2.Hardware) admission.Response {
	dups := duplicates{}
	for _, mac := range hw.GetMACs() {
		var hwWithMAC v1alpha2.HardwareList
		err := a.client.List(ctx, &hwWithMAC, ctrlclient.MatchingFields{
			internal.HardwareByMACAddr: mac,
		})
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}

		if len(hwWithMAC.Items) > 0 {
			dups.AppendTo(mac, hwWithMAC.Items...)
		}
	}

	if len(dups) > 0 {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf(
			"MAC associated with existing Hardware: %s",
			dups.String(),
		))
	}

	return admission.Allowed("")
}
