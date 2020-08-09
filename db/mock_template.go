package db

import (
	"context"

	"github.com/golang/protobuf/ptypes/timestamp"
	uuid "github.com/satori/go.uuid"
)

// CreateTemplate creates a new workflow template
func (mdb MockDB) CreateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {

	return nil
}

// GetTemplate returns a workflow template
func (mdb MockDB) GetTemplate(ctx context.Context, id string) (string, string, error) {

	return "", "", nil
}

// DeleteTemplate deletes a workflow template
func (mdb MockDB) DeleteTemplate(ctx context.Context, name string) error {

	return nil
}

// ListTemplates returns all saved templates
func (mdb MockDB) ListTemplates(fn func(id, n string, in, del *timestamp.Timestamp) error) error {

	return nil
}

// UpdateTemplate update a given template
func (mdb MockDB) UpdateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {

	return nil
}
