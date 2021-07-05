package template

import (
	"context"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/describe"
	"github.com/tinkerbell/tink/protos/template"
)

type describeTemplate struct {
	describe.Options
}

func (d *describeTemplate) DescribeByID(ctx context.Context, cl *client.FullClient, requestedID string) (interface{}, error) {
	return cl.TemplateClient.GetTemplate(ctx, &template.GetRequest{
		GetBy: &template.GetRequest_Id{
			Id: requestedID,
		},
	})
}

func NewDescribeOptions() describe.Options {
	t := describeTemplate{}
	return describe.Options{
		DescribeByID: t.DescribeByID,
	}
}
