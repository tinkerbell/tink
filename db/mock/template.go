package mock

import (
	"context"
	"errors"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
)

type template struct {
	id   uuid.UUID
	data string
}

var templateDB = map[string]interface{}{}

// CreateTemplate creates a new workflow template
func (d DB) CreateTemplate(ctx context.Context, name string, data string, id uuid.UUID) error {
	if len(templateDB) > 0 {
		if _, ok := templateDB[name]; ok {
			return errors.New("Template name already exist in the database")
		}
		templateDB[name] = template{
			id:   id,
			data: data,
		}

	} else {
		templateDB[name] = template{
			id:   id,
			data: data,
		}
	}
	return nil
}

// GetTemplate returns a workflow template
func (d DB) GetTemplate(ctx context.Context, id string) (string, string, error) {
	return d.GetTemplateFunc(ctx, id)
}

// DeleteTemplate deletes a workflow template
func (d DB) DeleteTemplate(ctx context.Context, name string) error {
	if len(templateDB) > 0 {
		if _, ok := templateDB[name]; !ok {
			return errors.New("Template name does not exist")
		}
		delete(templateDB, name)
	}
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

// ClearTemplateDB clear all the templates
func (d DB) ClearTemplateDB() {
	for name := range templateDB {
		delete(templateDB, name)
	}
}
