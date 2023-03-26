package internal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tinkerbell/tink/internal/agent/failure"
)

// NewFailureFiles creates a new FailureFiles instance with isolated underlying files. Consumers
// are responsible for calling FailureFiles.Close().
func NewFailureFiles() (*FailureFiles, error) {
	reasonFile, err := ioutil.TempFile("", "failure-reason-*")
	if err != nil {
		return nil, err
	}

	messageFile, err := ioutil.TempFile("", "failure-message-*")
	if err != nil {
		return nil, err
	}

	return &FailureFiles{
		reasonFile:  reasonFile,
		messageFile: messageFile,
	}, nil
}

// FailureFiles provides mountable files for runtimes that can be used to extract
// a reason and message from actions.
type FailureFiles struct {
	reasonFile  *os.File
	messageFile *os.File
}

// Close closes all files tracked by f.
func (f *FailureFiles) Close() error {
	os.Remove(f.reasonFile.Name())
	os.Remove(f.reasonFile.Name())
	return nil
}

// ReasonPath returns the path for the reason file.
func (f *FailureFiles) ReasonPath() string {
	return f.reasonFile.Name()
}

// Reason retrieves the reason from the reason file.
func (f *FailureFiles) Reason() (string, error) {
	var reason bytes.Buffer
	if _, err := reason.ReadFrom(f.reasonFile); err != nil {
		return "", err
	}
	return strings.TrimRight(reason.String(), "\n"), nil
}

// MessagePath returns the path for the message file.
func (f *FailureFiles) MessagePath() string {
	return f.messageFile.Name()
}

// Message retrieves the message from the message file.
func (f *FailureFiles) Message() (string, error) {
	var message bytes.Buffer
	if _, err := message.ReadFrom(f.messageFile); err != nil {
		return "", err
	}
	return strings.TrimRight(message.String(), "\n"), nil
}

func (f *FailureFiles) ToError() error {
	message, err := f.Message()
	if err != nil {
		return fmt.Errorf("read failure message: %w", err)
	}

	reason, err := f.Reason()
	if err != nil {
		return fmt.Errorf("read failure reason: %w", err)
	}

	return failure.WithReason(message, reason)
}
