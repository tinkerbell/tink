package grpcserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/protos/hardware"
)

func Test_server_normalizeHardwareData(t *testing.T) {
	tests := []struct {
		name     string
		given    *hardware.Hardware
		expected *hardware.Hardware
	}{
		{
			name: "Expect MAC to be normalized to lowercase",
			given: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:42:0E:D9:C7:53",
							},
						},
					},
				},
			},
			expected: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:42:0e:d9:c7:53",
							},
						},
					},
				},
			},
		},
		{
			name: "Expect MAC to be normalized to lowercase, multiple interfaces",
			given: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:42:0E:D9:C7:53",
							},
						},
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:4F:0E:D9:C7:5E",
							},
						},
					},
				},
			},
			expected: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:42:0e:d9:c7:53",
							},
						},
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:4f:0e:d9:c7:5e",
							},
						},
					},
				},
			},
		},
		{
			name: "Expect MAC to be normalized to lowercase, nultiple interfaces, mixed casing",
			given: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:42:0e:d9:c7:53",
							},
						},
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:4F:0E:D9:C7:5E",
							},
						},
					},
				},
			},
			expected: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:42:0e:d9:c7:53",
							},
						},
						{
							Dhcp: &hardware.Hardware_DHCP{
								Mac: "02:4f:0e:d9:c7:5e",
							},
						},
					},
				},
			},
		},
		{
			name: "Handle nil DHCP",
			given: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						{
							Dhcp: nil,
						},
					},
				},
			},
			expected: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						{
							Dhcp: nil,
						},
					},
				},
			},
		},
		{
			name: "Handle nil Interface",
			given: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						nil,
					},
				},
			},
			expected: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: []*hardware.Hardware_Network_Interface{
						nil,
					},
				},
			},
		},
		{
			name: "Handle nil Interfaces",
			given: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: nil,
				},
			},
			expected: &hardware.Hardware{
				Network: &hardware.Hardware_Network{
					Interfaces: nil,
				},
			},
		},
		{
			name:     "Handle nil Network",
			given:    &hardware.Hardware{Network: nil},
			expected: &hardware.Hardware{Network: nil},
		},
		{
			name:     "Handle nil Hardware",
			given:    nil,
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() { normalizeHardwareData(tt.given) })
			assert.Equal(t, tt.expected, tt.given)
		})
	}
}
