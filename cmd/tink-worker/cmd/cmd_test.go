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
  -id                   Worker ID.
  -log-level            Logging level. (default "info")
  -reg-pass             Container registry password.
  -reg-user             Container registry username.
  -registry             Container registry from which to pull images.
  -tink-cert-url        URL to Tink TLS certificate.
  -tink-grpc-authority  Hostname:port of Tink GRPC server.
`
	c := &Command{}
	fs := flag.NewFlagSet("tink-worker", flag.ExitOnError)
	cmd := newCLI(c, fs)
	out := cmd.UsageFunc(cmd)
	if diff := cmp.Diff(out, want); diff != "" {
		t.Fatal(diff)
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		cmd *Command
		err bool
	}{
		"success":          {cmd: &Command{ID: "0eba0bf8-3772-4b4a-ab9f-6ebe93b90a95", TinkCertURL: "https://tink.example.com/tink-cert.pem", TinkGRPCAuthority: "tink.example.com:443"}},
		"failure - bad ID": {cmd: &Command{ID: "asdf", TinkCertURL: "https://tink.example.com/tink-cert.pem", TinkGRPCAuthority: "tink.example.com:443"}, err: true},
		"failure - no ID":  {cmd: &Command{TinkCertURL: "https://tink.example.com/tink-cert.pem", TinkGRPCAuthority: "tink.example.com:443"}, err: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.cmd.Validate()
			if (got != nil) != tt.err {
				t.Fatalf("Command.Validate() error = %v, want %v", got, tt.err)
			}
		})
	}
}
