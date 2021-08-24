package template

import (
	"context"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/delete"
	"github.com/tinkerbell/tink/protos/template"
)

type deleteTemplate struct {
	delete.Options
}

func (d *deleteTemplate) DeleteByID(ctx context.Context, cl *client.FullClient, requestedID string) (interface{}, error) {
	return cl.TemplateClient.DeleteTemplate(ctx, &template.GetRequest{
		GetBy: &template.GetRequest_Id{
			Id: requestedID,
		},
	})
}

func NewDeleteOptions() delete.Options {
	t := deleteTemplate{}
	return delete.Options{
		DeleteByID: t.DeleteByID,
	}
}
