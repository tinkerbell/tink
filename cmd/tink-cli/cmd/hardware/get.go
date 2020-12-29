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
	get.CmdOpt
	metaClient *client.MetaClient
}

func NewGetHardwareOpt(metaClient *client.MetaClient) GetHardware {
	gh := GetHardware{
		CmdOpt: get.CmdOpt{
			Headers: []string{"ID", "MAC Address", "IP address", "Hostname"},
		},
		metaClient: metaClient,
	}
	gh.CmdOpt.PopulateTable = gh.PopulateTable
	gh.CmdOpt.RetrieveData = gh.RetrieveData
	return gh
}

func (h *GetHardware) RetrieveData(ctx context.Context) ([]interface{}, error) {
	list, err := h.metaClient.HardwareClient.All(ctx, &hardware.Empty{})
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
