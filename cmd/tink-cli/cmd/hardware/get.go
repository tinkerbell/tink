package hardware

import (
	"context"
	"io"

	"github.com/jedib0t/go-pretty/table"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	"github.com/tinkerbell/tink/protos/hardware"
)

type GetHardware struct {
	get.Options
	cl *client.FullClient
}

func NewGetHardware(cl *client.FullClient) GetHardware {
	gh := GetHardware{
		Options: get.Options{
			Headers: []string{"ID", "MAC Address", "IP address", "Hostname"},
		},
		cl: cl,
	}
	gh.Options.PopulateTable = gh.PopulateTable
	gh.Options.RetrieveData = gh.RetrieveData
	return gh
}

func (h *GetHardware) RetrieveData(ctx context.Context) ([]interface{}, error) {
	list, err := h.cl.HardwareClient.All(ctx, &hardware.Empty{})
	if err != nil {
		return nil, err
	}
	data := []interface{}{}
	var hw *hardware.Hardware
	for hw, err = list.Recv(); err == nil && hw != nil; hw, err = list.Recv() {
		data = append(data, hw)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return data, nil
}

func (h *GetHardware) PopulateTable(data []interface{}, t table.Writer) error {
	for _, v := range data {
		if hw, ok := v.(*hardware.Hardware); ok {
			// TODO(gianarb): I think we should
			// print it better. The hardware is one
			// even if if has more than one
			// interface.
			for _, iface := range hw.GetNetwork().GetInterfaces() {
				t.AppendRow(table.Row{hw.Id,
					iface.Dhcp.Mac,
					iface.Dhcp.Ip.Address,
					iface.Dhcp.Hostname})
			}
		}
	}
	return nil
}
