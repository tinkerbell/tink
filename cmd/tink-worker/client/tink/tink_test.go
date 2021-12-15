package tink

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/packethost/pkg/grpc"
	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/protos/workflow"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// Self-signed cert we use to test credential generation.
const selfSignedCert string = `-----BEGIN CERTIFICATE-----
MIIB0DCCATGgAwIBAgIBATAKBggqhkjOPQQDBDASMRAwDgYDVQQKEwdBY21lIENv
MB4XDTIxMTIwODIzNTYwMFoXDTIyMDYwNjIzNTYwMFowEjEQMA4GA1UEChMHQWNt
ZSBDbzCBmzAQBgcqhkjOPQIBBgUrgQQAIwOBhgAEAL71xJa5btFlW1BUaxrLAbsM
1tgi0UTB6w4CHm+6iwyuP0BZD9eMwSjMzhZ70lmlah0Z5Z8oHFHEFOIYZnIG4O/r
ATnzswVeVIPuX8uZ6ApsKq5q9uk9ByJIRfLhNqmrpLNSGbdcNEIoI27DIwiQdeT8
U6WYuWZB8BKBNre2Q/OiAcPNozUwMzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAww
CgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIwADAKBggqhkjOPQQDBAOBjAAwgYgCQgE1
AkrK0PoP1Mvia0jMGYBDwGzBCLjqPpPFuTE0pVJctog/ZXBRUDlSz9BN3PbCFJZs
F+/RUdtxQsgcBeBNrKk2qwJCAVxHlTzNK+hv7Xr3MXC39TopWJuNEj/1B37FZ8HG
0qv7ie1O3lcrMXd9evsIp9KLqSgUCyxmN2SS6LKBjaAwzmJd
-----END CERTIFICATE-----`

type fakeServer struct {
	workflow.WorkflowServiceServer
	data map[string]string
	err  error
}

const bufSize = 1024 * 1024

func startWorkflowServerAndConnectClient(t *testing.T, name string, server *grpc.Server) workflow.WorkflowServiceClient {
	t.Helper()

	ctx := context.Background()
	listener := bufconn.Listen(bufSize)
	go func() {
		t.Helper()

		if err := server.Server().Serve(listener); err != nil {
			t.Error(fmt.Errorf("%s.Serve exited with error: %w", name, err))
		}
	}()

	dialer := func(ctx context.Context, _ string) (net.Conn, error) {
		return listener.DialContext(ctx)
	}

	conn, err := ggrpc.DialContext(ctx, "bufnet", ggrpc.WithContextDialer(dialer), ggrpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial %s: %v", name, err)
	}

	client := workflow.NewWorkflowServiceClient(conn)
	return client
}

func TestObtainServerCredsWithInvalidCerts(t *testing.T) {
	emptyCert := ""
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, emptyCert)
	}))
	defer svr.Close()

	certTests := []struct {
		testName         string
		url              string
		expectedErrorMsg string
	}{
		{"certURL arg validation", "", "certURL cannot be empty"},
		{"ensure we get a response when fetching a cert", "something", "unable to fetch cert from something"},
		{"ensure an empty cert response is rejected", svr.URL, "unable to parse the cert data"},
	}
	for _, tt := range certTests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := ObtainServerCreds(tt.url)
			if err == nil {
				t.Fatal("Error expected but none found")
			} else if fmt.Sprint(err) != tt.expectedErrorMsg {
				t.Fatalf("Error should be %v, got %v", tt.expectedErrorMsg, err)
			}
		})
	}
}

func TestObtainServerCredsWithValidCert(t *testing.T) {
	t.Run("using with a self-signed cert", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, selfSignedCert)
		}))
		defer server.Close()

		_, err := ObtainServerCreds(server.URL)
		if err != nil {
			t.Fatalf("Error obtaining tink server creds: %v", err)
		}
	})
}

func TestEstablishServerConnection(t *testing.T) {
	t.Run("grpcAuthority arg validation", func(t *testing.T) {
		expectedErrorMsg := "grpcAuthority cannot be empty"
		_, err := EstablishServerConnection("", nil)
		if err == nil {
			t.Fatal("Error expected but none found")
		} else if fmt.Sprint(err) != expectedErrorMsg {
			t.Fatalf("Error should be %v, got %v", expectedErrorMsg, err)
		}
	})

	t.Run("attempt to connect to a bogus GRPC server", func(t *testing.T) {
		expectedErrorMsg := "connect to tinkerbell server"
		_, err := EstablishServerConnection("bogus", nil)
		if err == nil {
			t.Fatal("Error expected but none found")
		} else if fmt.Sprint(err) != expectedErrorMsg {
			t.Fatalf("Error should be %v, got %v", expectedErrorMsg, err)
		}
	})
}

func TestClientCreation(t *testing.T) {
	t.Run("ability to create a WorkflowServiceClient", func(t *testing.T) {
		name := "tinkWorkflowServer"
		fakeServer, err := grpc.NewServer(log.Test(t, name), func(s *grpc.Server) {
			workflow.RegisterWorkflowServiceServer(s.Server(), &fakeServer{
				data: nil,
				err:  nil,
			})
		})
		if err != nil {
			t.Fatalf("Error: %s", err)
		}
		if fakeServer == nil {
			t.Fatal("Error creating workflow server (server is nil)")
		}

		fakeClient := startWorkflowServerAndConnectClient(t, name, fakeServer)
		if fakeClient == nil {
			t.Fatal("Error creating workflow client (client is nil)")
		}
	})
}
