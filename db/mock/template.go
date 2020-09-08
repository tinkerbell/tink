package mock

import (
	"context"

	"github.com/golang/protobuf/ptypes/timestamp"
<<<<<<< HEAD
	"github.com/google/uuid"
	"github.com/tinkerbell/tink/pkg"
=======
	uuid "github.com/satori/go.uuid"
>>>>>>> Incorporating review comments
)

// CreateTemplate creates a new workflow template
func (d DB) CreateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {
	return d.CreateTemplateFunc(ctx, name, data, id)
}

// GetTemplate returns a workflow template
func (d DB) GetTemplate(ctx context.Context, id string) (string, string, error) {
	return d.GetTemplateFunc(ctx, id)
}

// DeleteTemplate deletes a workflow template
func (d DB) DeleteTemplate(ctx context.Context, name string) error {
	return nil
}

// ListTemplates returns all saved templates
func (d DB) ListTemplates(fn func(id, n string, in, del *timestamp.Timestamp) error) error {
	return nil
}

// UpdateTemplate update a given template
func (d DB) UpdateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {
	return nil
}
