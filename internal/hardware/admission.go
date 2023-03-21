package hardware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tinkerbell/tink/api/v1alpha2"
	"github.com/tinkerbell/tink/internal/hardware/internal"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// admissionWebhookEndpoint is the endpoint serving the Admission handler.
const admissionWebhookEndpoint = "/validate-tinkerbell-org-v1alpha2-hardware"

// +kubebuilder:webhook:path=/validate-tinkerbell-org-v1alpha2-hardware,mutating=false,failurePolicy=fail,groups="",resources=hardware,verbs=create;update,versions=v1alpha2,name=hardware.tinkerbell.org

// Admission handles complex validation for admitting a Hardware object to the cluster.
type Admission struct {
	client  ctrlclient.Client
	decoder *admission.Decoder
}

// Handle satisfies controller-runtime/pkg/webhook/admission#Handler. It is responsible for deciding
// if the given req is valid and should be admitted to the cluster.
func (a *Admission) Handle(ctx context.Context, req admission.Request) admission.Response {
	if a.client == nil {
		return admission.Errored(http.StatusInternalServerError, errors.New("misconfigured client"))
	}

	var hw v1alpha2.Hardware
	if err := a.decoder.Decode(req, &hw); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Ensure conditionally optional fields are valid
	if resp := a.validateConditionalFields(&hw); !resp.Allowed {
		return resp
	}

	// Ensure MACs on the hardware are valid.
	if resp := a.validateMACs(&hw); !resp.Allowed {
		return resp
	}

	// Ensure there's no hardware in the cluster with the same MAC addresses.
	if resp := a.validateUniqueMACs(ctx, &hw); !resp.Allowed {
		return resp
	}

	// Ensure there's no hardware in the cluster with the same IP addresses.
	if resp := a.validateUniqueIPs(ctx, &hw); !resp.Allowed {
		return resp
	}

	return admission.Allowed("")
}

// InjectDecoder satisfies controller-runtime/pkg/webhook/admission#DecoderInjector. It is used
// when registering the webhook to inject the decoder used by the controller manager.
func (a *Admission) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

// SetClient sets a's internal Kubernetes client.
func (a *Admission) SetClient(c ctrlclient.Client) {
	a.client = c
}

// SetupWithManager registers a with mgr as a webhook served from AdmissionWebhookEndpoint.
func (a *Admission) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	idx := mgr.GetFieldIndexer()

	err := idx.IndexField(
		ctx,
		&v1alpha2.Hardware{},
		internal.HardwareByMACAddr,
		internal.HardwareByMACAddrFunc,
	)
	if err != nil {
		return fmt.Errorf("register index %s: %w", internal.HardwareByMACAddr, err)
	}

	err = idx.IndexField(
		ctx,
		&v1alpha2.Hardware{},
		internal.HardwareByIPAddr,
		internal.HardwareByIPAddrFunc,
	)
	if err != nil {
		return fmt.Errorf("register index %s: %w", internal.HardwareByIPAddr, err)
	}

	mgr.GetWebhookServer().Register(
		admissionWebhookEndpoint,
		&webhook.Admission{Handler: a},
	)

	return nil
}
