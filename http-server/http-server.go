package httpserver

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net/http"
	"runtime"
	"time"

	grpcRuntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/target"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
)

var (
	gitRev         = "unknown"
	gitRevJSON     []byte
	grpcListenAddr = "localhost:42113"
	httpListenAddr = ":42114"
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
	err := hardware.RegisterHardwareServiceHandlerFromEndpoint(ctx, mux, grpcListenAddr, dialOpts)
	if err != nil {
		logger.Error(err)
	}
	err = target.RegisterTargetHandlerFromEndpoint(ctx, mux, grpcListenAddr, dialOpts)
	if err != nil {
		logger.Error(err)
	}
	err = template.RegisterTemplateHandlerFromEndpoint(ctx, mux, grpcListenAddr, dialOpts)
	if err != nil {
		logger.Error(err)
	}
	err = workflow.RegisterWorkflowSvcHandlerFromEndpoint(ctx, mux, grpcListenAddr, dialOpts)
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
	http.Handle("/", mux)

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
