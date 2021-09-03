package hardware

import (
	"context"
	"errors"
	"io"

	"github.com/jedib0t/go-pretty/table"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	hwpb "github.com/tinkerbell/tink/protos/hardware"
)

type getHardware struct {
	get.Options
}

func NewGetOptions() get.Options {
	gh := getHardware{}
	opt := get.Options{
		Headers:       []string{"ID", "MAC Address", "IP address", "Hostname"},
		PopulateTable: gh.PopulateTable,
		RetrieveData:  gh.RetrieveData,
		RetrieveByID:  gh.RetrieveByID,
	}
	return opt
}

func (h *getHardware) RetrieveByID(ctx context.Context, cl *client.FullClient, requiredID string) (interface{}, error) {
	return cl.HardwareClient.ByID(ctx, &hwpb.GetRequest{Id: requiredID})
}

func (h *getHardware) RetrieveData(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
	list, err := cl.HardwareClient.All(ctx, &hwpb.Empty{})
	if err != nil {
		return nil, err
	}
	data := []interface{}{}
	var hw *hwpb.Hardware
	for hw, err = list.Recv(); err == nil && hw != nil; hw, err = list.Recv() {
		data = append(data, hw)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return data, nil
}

func (h *getHardware) PopulateTable(data []interface{}, t table.Writer) error {
	for _, v := range data {
		if hw, ok := v.(*hwpb.Hardware); ok {
			// TODO(gianarb): I think we should
			// print it better. The hardware is one
			// even if if has more than one
			// interface.
			for _, iface := range hw.GetNetwork().GetInterfaces() {
				t.AppendRow(table.Row{
					hw.Id,
					iface.Dhcp.Mac,
					iface.Dhcp.Ip.Address,
					iface.Dhcp.Hostname,
				})
			}
		}
	}
	return nil
}
