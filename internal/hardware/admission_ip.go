package hardware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/tinkerbell/tink/api/v1alpha2"
	"github.com/tinkerbell/tink/internal/hardware/internal"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (a *Admission) validateUniqueIPs(ctx context.Context, hw *v1alpha2.Hardware) admission.Response {
	// Determine if there are IP duplicates within the hw object.
	seen := map[string]struct{}{}
	var dupOnHw []string
	for _, ip := range hw.GetIPs() {
		if _, ok := seen[ip]; ok {
			dupOnHw = append(dupOnHw, ip)
		}
		seen[ip] = struct{}{}
	}

	if len(dupOnHw) > 0 {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf(
			"duplicate IPs on Hardware: %v",
			strings.Join(dupOnHw, ", "),
		))
	}

	// Determine if there are IP duplicates with other Hardware objects.
	dups := duplicates{}
	for _, ip := range hw.GetIPs() {
		var hwWithIP v1alpha2.HardwareList
		err := a.client.List(ctx, &hwWithIP, ctrlclient.MatchingFields{
			internal.HardwareByIPAddr: ip,
		})
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		if len(hwWithIP.Items) > 0 {
			dups.AppendTo(ip, hwWithIP.Items...)
		}
	}

	if len(dups) > 0 {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf(
			"IP associated with existing Hardware: %v",
			dups.String(),
		))
	}

	return admission.Allowed("")
}
