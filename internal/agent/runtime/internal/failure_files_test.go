package internal_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/tinkerbell/tink/internal/agent/failure"
	"github.com/tinkerbell/tink/internal/agent/runtime/internal"
)

func TestFailureFiles(t *testing.T) {
	ff, err := internal.NewFailureFiles()
	if err != nil {
		t.Fatalf("Could not create failure files: %v", err)
	}

	expectMessage := "my special message"
	expectReason := "MyReason"

	fh, err := os.OpenFile(ff.MessagePath(), os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("Could not open message file: %v", err)
	}
	defer fh.Close()
	if _, err := io.WriteString(fh, expectMessage); err != nil {
		t.Fatalf("Couldn't write to message file: %v", err)
	}

	fh, err = os.OpenFile(ff.ReasonPath(), os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("Could not open reason file: %v", err)
	}
	defer fh.Close()
	if _, err := io.WriteString(fh, expectReason); err != nil {
		t.Fatalf("Couldn't write to reason file: %v", err)
	}

	// Read the individual messages and ensure they match.
	receivedMessage, err := ff.Message()
	if err != nil {
		t.Fatalf("Could not retrieve message: %v", err)
	}
	if receivedMessage != expectMessage {
		t.Fatalf("Expected: %v; Received: %v", expectMessage, receivedMessage)
	}

	receivedReason, err := ff.Reason()
	if err != nil {
		t.Fatalf("Could not retrieve message: %v", err)
	}
	if receivedReason != expectReason {
		t.Fatalf("Expected: %v; Received: %v", expectReason, receivedReason)
	}

	// Convert to an error and ensure we can extract using the failure package.
	toErr := ff.ToError()

	fmt.Printf("%T %v\n", toErr, toErr.Error())

	receivedReason, ok := failure.Reason(toErr)
	if !ok {
		t.Fatalf("Expected a reason that could be extracted with failure package, received none")
	}
	if receivedReason != expectReason {
		t.Fatalf("Expected: %v; Received: %v", receivedReason, expectReason)
	}

	if toErr.Error() != expectMessage {
		t.Fatalf("Expected: %v; Received: %v", expectMessage, toErr.Error())
	}

	// Close the files and ensure they've been deleted.
	ff.Close()

	_, err = os.Stat(ff.MessagePath())
	switch {
	case err == nil:
		t.Fatal("Expected os.Stat error but received none")
	case !os.IsNotExist(err):
		t.Fatalf("Expected not exists path error, received '%v'", err)
	}
}
