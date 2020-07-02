package httpserver

import (
	"bytes"
	"context"
	"crypto/subtle"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	grpcRuntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
)

var (
	gitRev         = "unknown"
	gitRevJSON     []byte
	grpcEndpoint   = os.Getenv("TINKERBELL_GRPC_AUTHORITY")
	httpListenAddr = os.Getenv("TINKERBELL_HTTP_AUTHORITY")
	authUsername   = os.Getenv("TINK_AUTH_USERNAME")
	authPassword   = os.Getenv("TINK_AUTH_PASSWORD")
	startTime      = time.Now()
	logger         log.Logger
)

// SetupHTTP setup and return an HTTP server
func SetupHTTP(ctx context.Context, lg log.Logger, certPEM []byte, modTime time.Time, errCh chan<- error) {
	logger = lg

	cp := x509.NewCertPool()
	ok := cp.AppendCertsFromPEM(certPEM)
	if !ok {
		logger.Error(errors.New("parse cert"))
	}

	creds := credentials.NewClientTLSFromCert(cp, "")

	mux := grpcRuntime.NewServeMux()

	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	if grpcEndpoint == "" {
		grpcEndpoint = "localhost:42113"
	}
	host, _, err := net.SplitHostPort(grpcEndpoint)
	if err != nil {
		logger.Error(err)
	}
	if host == "" {
		grpcEndpoint = "localhost" + grpcEndpoint
	}
	err = RegisterHardwareServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		logger.Error(err)
	}
	err = template.RegisterTemplateHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		logger.Error(err)
	}
	err = workflow.RegisterWorkflowSvcHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		logger.Error(err)
	}

	http.HandleFunc("/cert", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "server.pem", modTime, bytes.NewReader(certPEM))
	})
	http.Handle("/metrics", promhttp.Handler())
	setupGitRevJSON()
	http.HandleFunc("/version", versionHandler)
	http.HandleFunc("/_packet/healthcheck", healthCheckHandler)
	http.Handle("/", BasicAuth(mux))

	if httpListenAddr == "" {
		httpListenAddr = ":42114"
	}
	srv := &http.Server{
		Addr: httpListenAddr,
	}
	go func() {
		logger.Info("serving http")
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			err = nil
		}
		errCh <- err
	}()
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(gitRevJSON)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	res := struct {
		GitRev     string  `json:"git_rev"`
		Uptime     float64 `json:"uptime"`
		Goroutines int     `json:"goroutines"`
	}{
		GitRev:     gitRev,
		Uptime:     time.Since(startTime).Seconds(),
		Goroutines: runtime.NumGoroutine(),
	}

	b, err := json.Marshal(&res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func setupGitRevJSON() {
	res := struct {
		GitRev  string `json:"git_rev"`
		Service string `json:"service_name"`
	}{
		GitRev:  gitRev,
		Service: "tinkerbell",
	}
	b, err := json.Marshal(&res)
	if err != nil {
		err = errors.Wrap(err, "could not marshal version json")
		logger.Error(err)
		panic(err)
	}
	gitRevJSON = b
}

// BasicAuth adds authentication to the routes handled by handler
// skips authentication if both authUsername and authPassword aren't set
func BasicAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authUsername != "" || authPassword != "" {
			user, pass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(authUsername)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(authPassword)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Tink Realm"`)
				w.WriteHeader(401)
				w.Write([]byte("401 Unauthorized\n"))
				return
			}
		}
		handler.ServeHTTP(w, r)
	})
}
