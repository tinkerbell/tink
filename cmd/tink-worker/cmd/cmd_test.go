package cmd

import (
	"flag"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCustomUsageFunc(t *testing.T) {
	want := `USAGE
  Run Tink Worker

FLAGS
  -id         Worker ID.
  -log-level  Logging level. (default "info")
  -reg-pass   Container registry password.
  -reg-user   Container registry username.
  -registry   Container registry from which to pull images.
`
	c := &Command{}
	fs := flag.NewFlagSet("tink-worker", flag.ExitOnError)
	cmd := newCLI(c, fs)
	out := cmd.UsageFunc(cmd)
	if diff := cmp.Diff(out, want); diff != "" {
		t.Fatal(diff)
	}
}
