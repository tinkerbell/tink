package hardware_test

import (
	"context"
	"strings"
	"testing"

	"github.com/tinkerbell/tink/api/v1alpha2"
	"github.com/tinkerbell/tink/internal/hardware"
	"github.com/tinkerbell/tink/internal/hardware/internal"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	InvalidMAC    = "invalid MAC address"
	MACAssociated = "MAC associated with existing Hardware"
)

const (
	InvalidIP    = "invalid IP address"
	DuplicateIP  = "duplicate IPs on Hardware"
	IPAssociated = "IP associated with existing Hardware"
)

const DHCPEnabled = "DHCP enabled but no DHCP config"

func TestAdmissionHandler(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)

	// Configure the decoder for the Admission object.
	decoder := admission.NewDecoder(scheme)

	// Build the fake client with indexes so the Admission object can perform its lookups.
	// The indexes should be in sync with whatever indexes are registered via
	// hardware.Admission#SetupWithManager.
	cb := fake.NewClientBuilder().
		WithScheme(scheme).
		WithIndex(&v1alpha2.Hardware{}, internal.HardwareByMACAddr, internal.HardwareByMACAddrFunc).
		WithIndex(&v1alpha2.Hardware{}, internal.HardwareByIPAddr, internal.HardwareByIPAddrFunc)
	clnt := cb.Build()

	// Build the Admission object.
	adm := &hardware.Admission{}
	adm.SetClient(clnt)
	_ = adm.InjectDecoder(decoder)

	tests := []struct {
		Name             string
		Submission       *v1alpha2.Hardware
		Objects          []*v1alpha2.Hardware
		DisallowContains []string
	}{
		// Allowed
		{
			Name: "MultiInterfaces",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:00:00:00:00:00": v1alpha2.NetworkInterface{DisableDHCP: true},
						"00:00:00:00:00:01": v1alpha2.NetworkInterface{
							DHCP: &v1alpha2.DHCP{
								IP: "1.1.1.1",
							},
						},
						"00:00:00:00:00:02": v1alpha2.NetworkInterface{
							DHCP: &v1alpha2.DHCP{
								IP: "2.2.2.2",
							},
						},
					},
				},
			},
		},
		{
			Name: "ValidMAC1",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:00:00:00:00:00": v1alpha2.NetworkInterface{DisableDHCP: true},
					},
				},
			},
		},
		{
			Name: "ValidMAC2",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:12:34:56:78:90": v1alpha2.NetworkInterface{DisableDHCP: true},
					},
				},
			},
		},
		{
			Name: "ValidMAC3",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"aa:bb:cc:dd:ee:ff": v1alpha2.NetworkInterface{DisableDHCP: true},
					},
				},
			},
		},
		{
			Name: "MultiInterfaces",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:00:00:00:00:00": v1alpha2.NetworkInterface{DisableDHCP: true},
						"00:00:00:00:00:01": v1alpha2.NetworkInterface{
							DHCP: &v1alpha2.DHCP{
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
			Objects: []*v1alpha2.Hardware{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hw1",
					},
					Spec: v1alpha2.HardwareSpec{
						NetworkInterfaces: v1alpha2.NetworkInterfaces{
							"00:00:00:00:00:02": v1alpha2.NetworkInterface{
								DHCP: &v1alpha2.DHCP{
									IP: "2.2.2.2",
								},
							},
						},
					},
				},
			},
		},

		// Conditional fields
		{
			Name: "DHCPDisbaled",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"aa:bb:cc:dd:ee:ff": v1alpha2.NetworkInterface{DisableDHCP: true},
					},
				},
			},
		},
		{
			Name: "DHCPEnabledWithoutDHCP",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"aa:bb:cc:dd:ee:ff": v1alpha2.NetworkInterface{},
					},
				},
			},
			DisallowContains: []string{DHCPEnabled, "aa:bb:cc:dd:ee:ff"},
		},

		// Invalid MACs
		{
			Name: "EmptyMAC",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"": v1alpha2.NetworkInterface{DisableDHCP: true},
					},
				},
			},
			DisallowContains: []string{InvalidMAC, "empty"},
		},
		{
			Name: "ShortMAC",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:00:00:00:00:0": v1alpha2.NetworkInterface{
							DHCP: &v1alpha2.DHCP{
								IP:      "0.0.0.0",
								Netmask: "255.255.255.0",
							},
						},
					},
				},
			},
			DisallowContains: []string{InvalidMAC, "00:00:00:00:00:0"},
		},
		{
			Name: "UpperMAC",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"AA:BB:CC:DD:EE:FF": v1alpha2.NetworkInterface{
							DHCP: &v1alpha2.DHCP{
								IP:      "0.0.0.0",
								Netmask: "255.255.255.0",
							},
						},
					},
				},
			},
			DisallowContains: []string{InvalidMAC, "AA:BB:CC:DD:EE:FF"},
		},
		{
			Name: "MultiInvalidMAC",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"":               v1alpha2.NetworkInterface{DisableDHCP: true},
						"00:00:00":       v1alpha2.NetworkInterface{DisableDHCP: true},
						"11:00:11:00":    v1alpha2.NetworkInterface{DisableDHCP: true},
						"AA:00:00:00:00": v1alpha2.NetworkInterface{DisableDHCP: true},
					},
				},
			},
			DisallowContains: []string{
				InvalidMAC,
				"empty",
				"00:00:00",
				"11:00:11:00",
				"AA:00:00:00:00",
			},
		},

		// MAC duplication
		{
			Name: "DuplicateMAC",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:00:00:00:00:00": v1alpha2.NetworkInterface{DisableDHCP: true},
					},
				},
			},
			Objects: []*v1alpha2.Hardware{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hw1",
					},
					Spec: v1alpha2.HardwareSpec{
						NetworkInterfaces: v1alpha2.NetworkInterfaces{
							"00:00:00:00:00:00": v1alpha2.NetworkInterface{DisableDHCP: true},
						},
					},
				},
			},
			DisallowContains: []string{MACAssociated, "hw1", "00:00:00:00:00:00"},
		},
		{
			Name: "MACAssociated",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:00:00:00:00:00": v1alpha2.NetworkInterface{DisableDHCP: true},
					},
				},
			},
			Objects: []*v1alpha2.Hardware{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hw1",
					},
					Spec: v1alpha2.HardwareSpec{
						NetworkInterfaces: v1alpha2.NetworkInterfaces{
							"00:00:00:00:00:00": v1alpha2.NetworkInterface{DisableDHCP: true},
						},
					},
				},
			},
			DisallowContains: []string{MACAssociated, "hw1", "00:00:00:00:00:00"},
		},

		// IP duplication
		{
			Name: "DuplicateIP",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:00:00:00:00:00": v1alpha2.NetworkInterface{
							DHCP: &v1alpha2.DHCP{
								IP: "1.1.1.1",
							},
						},
						"00:00:00:00:00:01": v1alpha2.NetworkInterface{
							DHCP: &v1alpha2.DHCP{
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
			DisallowContains: []string{DuplicateIP, "1.1.1.1"},
		},
		{
			Name: "IPAssociated",
			Submission: &v1alpha2.Hardware{
				Spec: v1alpha2.HardwareSpec{
					NetworkInterfaces: v1alpha2.NetworkInterfaces{
						"00:00:00:00:00:00": v1alpha2.NetworkInterface{
							DHCP: &v1alpha2.DHCP{
								IP: "1.1.1.1",
							},
						},
					},
				},
			},
			Objects: []*v1alpha2.Hardware{
				{
					Spec: v1alpha2.HardwareSpec{
						NetworkInterfaces: v1alpha2.NetworkInterfaces{
							"00:00:00:00:00:01": v1alpha2.NetworkInterface{
								DHCP: &v1alpha2.DHCP{
									IP: "1.1.1.1",
								},
							},
						},
					},
				},
			},
			DisallowContains: []string{IPAssociated, "1.1.1.1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			// Clear out all objects from previous tests and register the new ones.
			if err := clnt.DeleteAllOf(context.Background(), &v1alpha2.Hardware{}); err != nil {
				t.Fatalf("delete existing objects: %v", err)
			}
			for _, o := range tc.Objects {
				// If the object doesn't have a name fill it in.
				if o.Name == "" {
					o.Name = rand.String(10)
				}
				if err := clnt.Create(context.Background(), o); err != nil {
					t.Fatalf("registering objects with fake client: %v", err)
				}
			}

			// We're assuming the json marshaller works with the controller runtime decoder.
			buf, err := json.Marshal(tc.Submission)
			if err != nil {
				t.Fatalf("encoding test object: %v", err)
			}

			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: buf,
					},
				},
			}

			// Run the object through the handler.
			resp := adm.Handle(context.Background(), req)

			if len(tc.DisallowContains) == 0 {
				if !resp.Allowed {
					t.Fatalf("disallowed: %v", resp.Result.Message)
				}
			} else {
				if resp.Allowed {
					t.Fatalf("expected object to be disallowed but was allowed")
				}

				for _, substr := range tc.DisallowContains {
					if !strings.Contains(resp.Result.Message, substr) {
						t.Fatalf(
							"expected reason to contain '%v' but got '%v'",
							substr,
							resp.Result.Message,
						)
					}
				}
			}
		})
	}
}
