// Copyright © 2018 packet.net

package hardware

import (
	"context"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/delete"
	hwpb "github.com/tinkerbell/tink/protos/hardware"
)

type deleteHardware struct {
	delete.Options
}

func (h *deleteHardware) DeleteByID(ctx context.Context, cl *client.FullClient, requestedID string) (interface{}, error) {
	return cl.HardwareClient.Delete(ctx, &hwpb.DeleteRequest{Id: requestedID})
}

func NewDeleteOptions() delete.Options {
	h := deleteHardware{}
	return delete.Options{
		DeleteByID: h.DeleteByID,
	}
}
