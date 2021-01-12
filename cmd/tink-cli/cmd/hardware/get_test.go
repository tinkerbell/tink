package hardware

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	hardware_proto "github.com/tinkerbell/tink/protos/hardware"
	"google.golang.org/grpc"
)

func TestGetHardware(t *testing.T) {
	table := []struct {
		counter           int
		Name              string
		ReturnedHardwares []*hardware_proto.Hardware
		Args              []string
		ExpectedStdout    string
	}{
		{
			Name: "happy-path",
			ReturnedHardwares: []*hardware_proto.Hardware{
				{
					Network: &hardware_proto.Hardware_Network{
						Interfaces: []*hardware_proto.Hardware_Network_Interface{
							{
								Dhcp: &hardware_proto.Hardware_DHCP{
									Mac: "cb:26:ad:28:5a:72",
									Ip: &hardware_proto.Hardware_DHCP_IP{
										Address: "192.168.12.12",
									},
									Hostname: "unittest-golang",
								},
							},
						},
					},
					Id: "1234-test",
				},
			},
			ExpectedStdout: `+-----------+-------------------+---------------+-----------------+
| ID        | MAC ADDRESS       | IP ADDRESS    | HOSTNAME        |
+-----------+-------------------+---------------+-----------------+
| 1234-test | cb:26:ad:28:5a:72 | 192.168.12.12 | unittest-golang |
+-----------+-------------------+---------------+-----------------+
`,
		},
		{
			Name: "get-json",
			Args: []string{"--format", "json"},
			ReturnedHardwares: []*hardware_proto.Hardware{
				{
					Network: &hardware_proto.Hardware_Network{
						Interfaces: []*hardware_proto.Hardware_Network_Interface{
							{
								Dhcp: &hardware_proto.Hardware_DHCP{
									Mac: "cb:26:ad:28:5a:72",
									Ip: &hardware_proto.Hardware_DHCP_IP{
										Address: "192.168.12.12",
									},
									Hostname: "unittest-golang",
								},
							},
						},
					},
					Id: "1234-test",
				},
			},
			ExpectedStdout: `{"data":[{"network":{"interfaces":[{"dhcp":{"mac":"cb:26:ad:28:5a:72","hostname":"unittest-golang","ip":{"address":"192.168.12.12"}}}]},"id":"1234-test"}]}`,
		},
	}

	for _, s := range table {
		t.Run(s.Name, func(t *testing.T) {
			cl := &client.FullClient{
				HardwareClient: &hardware_proto.HardwareServiceClientMock{
					AllFunc: func(ctx context.Context, in *hardware_proto.Empty, opts ...grpc.CallOption) (hardware_proto.HardwareService_AllClient, error) {
						return &hardware_proto.HardwareService_AllClientMock{
							RecvFunc: func() (*hardware_proto.Hardware, error) {
								s.counter = s.counter + 1
								if s.counter > len(s.ReturnedHardwares) {
									return nil, io.EOF
								}
								return s.ReturnedHardwares[s.counter-1], nil
							},
						}, nil
					},
				},
			}
			stdout := bytes.NewBufferString("")
			cmd := get.NewGetCommand(NewGetHardware(cl).Options)
			cmd.SetOut(stdout)
			cmd.SetArgs(s.Args)
			err := cmd.Execute()
			if err != nil {
				t.Error(err)
			}
			out, err := ioutil.ReadAll(stdout)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(string(out), s.ExpectedStdout); diff != "" {
				t.Error(diff)
			}
		})
	}
}
