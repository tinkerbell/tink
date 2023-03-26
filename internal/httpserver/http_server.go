package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	gitRev    = "unknown"
	startTime = time.Now()
	logger    logr.Logger
)

// SetupHTTP setup and return an HTTP server.
func SetupHTTP(ctx context.Context, logger logr.Logger, authority string, errCh chan<- error) {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/version", getGitRevJSONHandler())
	http.HandleFunc("/healthz", healthCheckHandler)

	srv := &http.Server{ //nolint:gosec // TODO: fix Potential Slowloris Attack because ReadHeaderTimeout is not configured
		Addr: authority,
	}
	go func() {
		logger.Info("serving http")
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()
	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error(err, "shutting down http server")
		}
	}()
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
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

func getGitRevJSONHandler() http.HandlerFunc {
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
		logger.Error(err, "")
		panic(err)
	}

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b)
	}
}
