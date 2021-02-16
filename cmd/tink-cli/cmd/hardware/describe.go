package hardware

import (
	"context"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/describe"
	"github.com/tinkerbell/tink/protos/hardware"
)

type describeHardware struct {
	describe.Options
}

func NewDescribeOptions() describe.Options {
	gh := describeHardware{}
	opt := describe.Options{
		DescribeByID: gh.DescribeByID,
	}
	return opt
}

func (h *describeHardware) DescribeByID(ctx context.Context, cl *client.FullClient, requiredID string) (interface{}, error) {
	return cl.HardwareClient.ByID(ctx, &hardware.GetRequest{Id: requiredID})
}
