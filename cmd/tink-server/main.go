package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/internal/grpcserver"
	"github.com/tinkerbell/tink/internal/httpserver"
	"github.com/tinkerbell/tink/internal/server"
	"go.uber.org/zap"
)

// version is set at build time.
var version = "devel"

// Config represents all the values you can configure as part of the tink-server.
// You can change the configuration via environment variable, or file, or command flags.
type Config struct {
	GRPCAuthority string
	HTTPAuthority string
	Backend       string

	KubeconfigPath string
	KubeAPI        string
	KubeNamespace  string
}

const backendKubernetes = "kubernetes"

func backends() []string {
	return []string{backendKubernetes}
}

func (c *Config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.GRPCAuthority, "grpc-authority", ":42113", "The address used to expose the gRPC server")
	fs.StringVar(&c.HTTPAuthority, "http-authority", ":42114", "The address used to expose the HTTP server")
	fs.StringVar(&c.Backend, "backend", backendKubernetes, fmt.Sprintf("The backend datastore to use. Must be one of %s", strings.Join(backends(), ", ")))
	fs.StringVar(&c.KubeconfigPath, "kubeconfig", "", "The path to the Kubeconfig. Only takes effect if `--backend=kubernetes`")
	fs.StringVar(&c.KubeAPI, "kubernetes", "", "The Kubernetes API URL, used for in-cluster client construction. Only takes effect if `--backend=kubernetes`")
	fs.StringVar(&c.KubeNamespace, "kube-namespace", "", "The Kubernetes namespace to target")
}

func (c *Config) PopulateFromLegacyEnvVar() {
	if v, ok := os.LookupEnv("TINKERBELL_GRPC_AUTHORITY"); ok {
		c.GRPCAuthority = v
	}

	if v, ok := os.LookupEnv("TINKERBELL_HTTP_AUTHORITY"); ok {
		c.HTTPAuthority = v
	}
}

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	var config Config

	zlog, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger := zapr.NewLogger(zlog).WithName("github.com/tinkerbell/tink")

	cmd := &cobra.Command{
		Use: "tink-server",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			viper, err := createViper(logger)
			if err != nil {
				return err
			}
			return applyViper(viper, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// I am not sure if it is right for this to be here,
			// but as last step I want to keep compatibility with
			// what we have for a little bit and I thinik that's
			// the most aggressive way we have to guarantee that
			// the old way works as before.
			config.PopulateFromLegacyEnvVar()

			logger.Info("Starting version " + version)

			ctx, oshutdown := otelinit.InitOpenTelemetry(cmd.Context(), "github.com/tinkerbell/tink")
			defer oshutdown(context.Background())

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
			ctx, closer := context.WithCancel(ctx)
			defer closer()
			// TODO(gianarb): I think we can do better in terms of
			// graceful shutdown and error management but I want to
			// figure this out in another PR
			errCh := make(chan error, 2)
			var registrar grpcserver.Registrar

			switch config.Backend {
			case backendKubernetes:
				var err error
				registrar, err = server.NewKubeBackedServer(
					logger,
					config.KubeconfigPath,
					config.KubeAPI,
					config.KubeNamespace,
				)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid backend: %s", config.Backend)
			}

			// Start the gRPC server in the background
			addr, err := grpcserver.SetupGRPC(
				ctx,
				registrar,
				config.GRPCAuthority,
				errCh,
			)
			if err != nil {
				return err
			}
			logger.Info("started listener", "address", addr)

			httpserver.SetupHTTP(ctx, logger, config.HTTPAuthority, errCh)

			select {
			case err := <-errCh:
				logger.Error(err, "")
			case sig := <-sigs:
				logger.Info("signal received, stopping servers", "signal", sig.String())
				closer()
			}

			// wait for grpc server to shutdown
			err = <-errCh
			if err != nil {
				return err
			}
			err = <-errCh
			if err != nil {
				return err
			}
			return nil
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}

func createViper(log logr.Logger) (*viper.Viper, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigName("tink-server")
	v.AddConfigPath("/etc/tinkerbell")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("could not load config file: %w", err)
		}
		log.Info("No config file found")
	} else {
		log.Info("Loaded config file", "path", v.ConfigFileUsed())
	}

	return v, nil
}

func applyViper(v *viper.Viper, cmd *cobra.Command) error {
	errors := []error{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				errors = append(errors, err)
				return
			}
		}
	})

	if len(errors) > 0 {
		errs := []string{}
		for _, err := range errors {
			errs = append(errs, err.Error())
		}
		return fmt.Errorf(strings.Join(errs, ", "))
	}

	return nil
}
