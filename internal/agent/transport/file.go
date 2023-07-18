package transport

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/workflow"
	"gopkg.in/yaml.v2"
)

// File is a transport implementation that executes a single workflow stored as a file.
type File struct {
	// Log is a logger for debugging.
	Log logr.Logger

	// Path to the workflow to run.
	Path string
}

// Start begins watching f.Dir for files. When it finds a file it hasn't handled before, it
// attempts to parse it and offload to the handler. It will run workflows once where a workflow
// is determined by its file name.
func (f *File) Start(ctx context.Context, _ string, handler WorkflowHandler) error {
	path, err := filepath.Abs(f.Path)
	if err != nil {
		return err
	}

	fh, err := os.Open(path)
	if err != nil {
		return err
	}

	var wrkflow workflow.Workflow
	if err := yaml.NewDecoder(fh).Decode(&wrkflow); err != nil {
		return err
	}

	handler.HandleWorkflow(ctx, wrkflow, f)

	return nil
}

func (f *File) RecordEvent(_ context.Context, e event.Event) error {
	// Noop because we don't particularly care about events for File based transports. Maybe
	// we'll record this in a dedicated file one day.
	f.Log.Info("Recording event", "event", e.GetName())
	return nil
}
