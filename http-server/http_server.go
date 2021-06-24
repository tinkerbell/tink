package httpserver

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net"
	"net/http"
	"runtime"
	"time"

	grpcRuntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	gitRev     = "unknown"
	gitRevJSON []byte
	startTime  = time.Now()
	logger     log.Logger
)

type HTTPServerConfig struct {
	CACertPath            string
	GRPCAuthority         string
	HTTPAuthority         string
	HTTPBasicAuthUsername string
	HTTPBasicAuthPassword string
}

// SetupHTTP setup and return an HTTP server
func SetupHTTP(ctx context.Context, lg log.Logger, config *HTTPServerConfig, errCh chan<- error) error {
	logger = lg

	creds, err := credentials.NewClientTLSFromFile(config.CACertPath, "")
	if err != nil {
		return err
	}

	mux := grpcRuntime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	grpcEndpoint := config.GRPCAuthority

	host, _, err := net.SplitHostPort(grpcEndpoint)
	if err != nil {
		return err
	}
	if host == "" {
		grpcEndpoint = "localhost" + grpcEndpoint
	}
	if err := RegisterHardwareServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts); err != nil {
		return err
	}
	if err := RegisterTemplateHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts); err != nil {
		return err
	}
	if err := RegisterWorkflowSvcHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts); err != nil {
		return err
	}

	http.HandleFunc("/cert", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, config.CACertPath)
	})
	http.Handle("/metrics", promhttp.Handler())
	setupGitRevJSON()
	http.HandleFunc("/version", versionHandler)
	http.HandleFunc("/healthz", healthCheckHandler)
	http.Handle("/", BasicAuth(config.HTTPBasicAuthUsername, config.HTTPBasicAuthPassword, mux))

	srv := &http.Server{
		Addr: config.HTTPAuthority,
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
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error(err)
		}
	}()

	return nil
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(gitRevJSON)
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
	_, _ = w.Write(b)
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
func BasicAuth(authUsername, authPassword string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authUsername != "" || authPassword != "" {
			user, pass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(authUsername)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(authPassword)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Tink Realm"`)
				w.WriteHeader(401)
				_, _ = w.Write([]byte("401 Unauthorized\n"))
				return
			}
		}
		handler.ServeHTTP(w, r)
	})
}
