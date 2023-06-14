package internal

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/tinkerbell/tink/internal/agent/failure"
)

// NewFailureFiles creates a new FailureFiles instance with isolated underlying files. Consumers
// are responsible for calling FailureFiles.Close().
func NewFailureFiles() (*FailureFiles, error) {
	reason, err := os.CreateTemp("", "failure-reason-*")
	if err != nil {
		return nil, err
	}

	message, err := os.CreateTemp("", "failure-message-*")
	if err != nil {
		return nil, err
	}

	return &FailureFiles{
		reason:  reason,
		message: message,
	}, nil
}

// FailureFiles provides mountable files for runtimes that can be used to extract
// a reason and message from actions.
type FailureFiles struct {
	reason  *os.File
	message *os.File
}

// Close closes all files tracked by f.
func (f *FailureFiles) Close() error {
	os.Remove(f.reason.Name())
	os.Remove(f.message.Name())
	return nil
}

// ReasonPath returns the path for the reason file.
func (f *FailureFiles) ReasonPath() string {
	return f.reason.Name()
}

// Reason retrieves the reason from the reason file.
func (f *FailureFiles) Reason() (string, error) {
	// Always seek back to the original point. If this fails, assume the file is missing and so
	// any further interactions will also receive errors.
	defer func() {
		_, _ = f.reason.Seek(0, 0)
	}()

	var reason bytes.Buffer
	if _, err := reason.ReadFrom(f.reason); err != nil {
		return "", err
	}
	return strings.TrimRight(reason.String(), "\n"), nil
}

// MessagePath returns the path for the message file.
func (f *FailureFiles) MessagePath() string {
	return f.message.Name()
}

// Message retrieves the message from the message file.
func (f *FailureFiles) Message() (string, error) {
	// Always seek back to the original point. If this fails, assume the file is missing and so
	// any further interactions will also receive errors.
	defer func() {
		_, _ = f.message.Seek(0, 0)
	}()

	var message bytes.Buffer
	if _, err := message.ReadFrom(f.message); err != nil {
		return "", err
	}
	return strings.TrimRight(message.String(), "\n"), nil
}

func (f *FailureFiles) ToError() error {
	// Always seek back to the original point. If this fails, assume the file is missing and so
	// any further interactions will also receive errors.
	defer func() {
		_, _ = f.reason.Seek(0, 0)
		_, _ = f.message.Seek(0, 0)
	}()

	message, err := f.Message()
	if err != nil {
		return fmt.Errorf("read failure message: %w", err)
	}

	reason, err := f.Reason()
	if err != nil {
		return fmt.Errorf("read failure reason: %w", err)
	}

	return failure.NewReason(message, reason)
}
