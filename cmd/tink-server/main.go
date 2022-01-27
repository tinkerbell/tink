package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/packethost/pkg/env"
	"github.com/packethost/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/cmd/tink-server/internal"
	grpcserver "github.com/tinkerbell/tink/grpc-server"
	httpserver "github.com/tinkerbell/tink/http-server"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// version is set at build time.
var version = "devel"

// DaemonConfig represents all the values you can configure as part of the tink-server.
// You can change the configuration via environment variable, or file, or command flags.
type DaemonConfig struct {
	Facility      string
	PGDatabase    string
	PGUSer        string
	PGPassword    string
	PGSSLMode     string
	OnlyMigration bool
	GRPCAuthority string
	TLSCert       string
	CertDir       string
	HTTPAuthority string
	TLS           bool
}

func (c *DaemonConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.Facility, "facility", "deprecated", "This is temporary. It will be removed")
	fs.StringVar(&c.PGDatabase, "postgres-database", "tinkerbell", "The Postgres database name")
	fs.StringVar(&c.PGUSer, "postgres-user", "tinkerbell", "The Postgres database username")
	fs.StringVar(&c.PGPassword, "postgres-password", "tinkerbell", "The Postgres database password")
	fs.StringVar(&c.PGSSLMode, "postgres-sslmode", "disable", "Enable or disable SSL mode in postgres")
	fs.BoolVar(&c.OnlyMigration, "only-migration", false, "When enabled the server applies the migration to postgres database and it exits")
	fs.StringVar(&c.GRPCAuthority, "grpc-authority", ":42113", "The address used to expose the gRPC server")
	fs.StringVar(&c.TLSCert, "tls-cert", "", "")
	fs.StringVar(&c.CertDir, "cert-dir", "", "")
	fs.StringVar(&c.HTTPAuthority, "http-authority", ":42114", "The address used to expose the HTTP server")
	fs.BoolVar(&c.TLS, "tls", true, "Run in tls protected mode (disabling should only be done for development or if behind TLS terminating proxy)")
}

func (c *DaemonConfig) PopulateFromLegacyEnvVar() {
	c.Facility = env.Get("FACILITY", c.Facility)

	c.PGDatabase = env.Get("PGDATABASE", c.PGDatabase)
	c.PGUSer = env.Get("PGUSER", c.PGUSer)
	c.PGPassword = env.Get("PGPASSWORD", c.PGPassword)
	c.PGSSLMode = env.Get("PGSSLMODE", c.PGSSLMode)
	c.OnlyMigration = env.Bool("ONLY_MIGRATION", c.OnlyMigration)

	c.TLSCert = env.Get("TINKERBELL_TLS_CERT", c.TLSCert)
	c.CertDir = env.Get("TINKERBELL_CERTS_DIR", c.CertDir)
	c.GRPCAuthority = env.Get("TINKERBELL_GRPC_AUTHORITY", c.GRPCAuthority)
	c.HTTPAuthority = env.Get("TINKERBELL_HTTP_AUTHORITY", c.HTTPAuthority)
	c.TLS = env.Bool("TINKERBELL_TLS", c.TLS)
}

func main() {
	logger, err := log.Init("github.com/tinkerbell/tink")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx, otelShutdown := otelinit.InitOpenTelemetry(ctx, "github.com/tinkerbell/tink")

	config := &DaemonConfig{}

	cmd := NewRootCommand(config, logger)
	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}

	logger.Close()
	otelShutdown(ctx)
}

func NewRootCommand(config *DaemonConfig, logger log.Logger) *cobra.Command {
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
			metrics.SetupMetrics(config.Facility, logger)

			logger.Info("starting version " + version)

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
			ctx, closer := context.WithCancel(cmd.Context())
			defer closer()
			// TODO(gianarb): I think we can do better in terms of
			// graceful shutdown and error management but I want to
			// figure this out in another PR
			errCh := make(chan error, 2)

			// TODO(gianarb): I moved this up because we need to be sure that both
			// connection, the one used for the resources and the one used for
			// listening to events and notification are coming in the same way.
			// BUT we should be using the right flags
			connInfo := fmt.Sprintf("dbname=%s user=%s password=%s sslmode=%s",
				config.PGDatabase,
				config.PGUSer,
				config.PGPassword,
				config.PGSSLMode,
			)
			database, err := internal.SetupPostgres(connInfo, config.OnlyMigration, logger)
			if err != nil {
				return err
			}
			if config.OnlyMigration {
				return nil
			}

			var (
				grpcOpts    []grpc.ServerOption
				certPEM     []byte
				certModTime *time.Time
			)
			if config.TLS {
				certsDir := os.Getenv("TINKERBELL_CERTS_DIR")
				if certsDir == "" {
					certsDir = filepath.Join("/certs", config.Facility)
				}
				var cert *tls.Certificate
				cert, certPEM, certModTime, err = grpcserver.GetCerts(certsDir)
				if err != nil {
					return err
				}
				grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewServerTLSFromCert(cert)))
			}

			tinkAPI, err := server.NewDBServer(
				logger,
				database,
				server.WithCerts(*certModTime, certPEM),
			)
			if err != nil {
				return err
			}

			// Start the gRPC server in the background
			addr, err := grpcserver.SetupGRPC(
				ctx,
				tinkAPI,
				config.GRPCAuthority,
				grpcOpts,
				errCh)
			if err != nil {
				return err
			}
			logger.With("address", addr).Info("started listener")

			httpConfig := &httpserver.Config{
				HTTPAuthority: config.HTTPAuthority,
			}
			httpserver.SetupHTTP(ctx, logger, httpConfig, errCh)

			select {
			case err = <-errCh:
				logger.Error(err)
			case sig := <-sigs:
				logger.With("signal", sig.String()).Info("signal received, stopping servers")
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

func createViper(logger log.Logger) (*viper.Viper, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigName("tink-server")
	v.AddConfigPath("/etc/tinkerbell")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.With("configFile", v.ConfigFileUsed()).Error(err, "could not load config file")
			return nil, err
		}
		logger.Info("no config file found")
	} else {
		logger.With("configFile", v.ConfigFileUsed()).Info("loaded config file")
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
